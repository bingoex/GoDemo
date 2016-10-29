package attrapi

// implementaton in pure go.
//
// Note: No need to detach memeory since shm willbe automatically detached en
// process _exit(), refter to man(2) shmdt for more info

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

const (
	KSumShmKey = 0x1111
	KSetShmKey = 0x2222
	KIpcCreate = 1 << 9
	KIpcExcl   = 2 << 9

	kNodeSize = 8 // AttrNode有8个字节

	kHashShmLen   = 1000 // shm一阶的大小，即一阶shm有1000个AttrNode
	kHashShmTimes = 40   // shm 阶数

	kNodeNum = kHashShmLen * kHashShmTimes
	kShmSize = kNodeSize * kNodeNum

	kMaxAttemp = 50
)

var hashModes = [kHashShmTimes]uint32{
	998, 997, 991, 983, 982, 977, 976, 974, 971, 967,
	964, 958, 956, 953, 947, 944, 941, 937, 934, 932,
	929, 926, 922, 919, 916, 914, 911, 908, 907, 904,
	898, 892, 887, 886, 883, 881, 878, 877, 872, 866,
}

var (
	setShm *uint8 // save attached set shm ptr
	sumShm *uint8 // save attached sum shm ptr
)

type node uint64

func (n node) String() string {
	return fmt.Sprintf("id: %d, val: %d", n.ID(), n.Val())
}

func (n *node) ID() uint32 {
	return *(*uint32)(unsafe.Pointer(n))
}

func (n *node) Val() uint32 {
	pVal := (uintptr)(unsafe.Pointer(n)) + 4
	return *(*uint32)(unsafe.Pointer(pVal))
}

func (n *node) Uint64Ptr() *uint64 {
	return (*uint64)(unsafe.Pointer(n))
}

func (n node) Uint64Val() uint64 {
	return uint64(n)
}

func makeNode(id, val uint32) node {
	var result uint64
	result = uint64(val)
	result = result << 32
	result = result | uint64(id)

	return node(result)
}

func newNode(id, val uint32) *node {
	n := makeNode(id, val)
	return &n
}

func attachShm(key, size int) (addr *uint8, err error) {
	var (
		shmid    uintptr
		shmAddr  uintptr
		sysErr   syscall.Errno
		shmExist bool = true
	)

	shmid, _, sysErr = syscall.RawSyscall(syscall.SYS_SHMGET,
		uintptr(key), uintptr(size), uintptr(0666))
	if sysErr != 0 {
		shmExist = false
		// try again with KIpcCreate
		shmid, _, sysErr = syscall.RawSyscall(syscall.SYS_SHMGET,
			uintptr(key), uintptr(size), uintptr(0666|KIpcCreate))
		if sysErr != 0 {
			return nil, errors.New(sysErr.Error())
		}
	}

	shmAddr, _, sysErr = syscall.RawSyscall(syscall.SYS_SHMAT, uintptr(shmid), 0, 0)
	if sysErr != 0 {
		return nil, errors.New(sysErr.Error())
	}

	addr = (*uint8)(unsafe.Pointer(shmAddr))

	// memset to zero
	if false == shmExist {
		c := (*[kNodeNum]uint64)(unsafe.Pointer(shmAddr))
		for i := 0; i < kShmSize/kNodeSize; i++ {
			c[i] = 0x0
		}
	}

	return addr, nil
}

func AttachAttrShm() (setShm, sumShm *uint8, err error) {
	shmSize := kNodeSize * kHashShmLen * kHashShmTimes
	if setShm, err = attachShm(KSetShmKey, shmSize); err != nil {
		fmt.Println("attach set shm failed", err)
		return nil, nil, err
	}

	if sumShm, err = attachShm(KSumShmKey, shmSize); err != nil {
		fmt.Println("attach sum shm failed", err)
		syscall.RawSyscall(syscall.SYS_SHMDT, uintptr((unsafe.Pointer(setShm))), 0, 0)
		return nil, nil, err
	}

	return setShm, sumShm, nil
}

var one sync.Once

// Attach shm at startup
func init() {
	log.SetFlags(log.Lshortfile)
	one.Do(func() {
		var err error
		if setShm, sumShm, err = AttachAttrShm(); err != nil {
			log.Fatal("!AttachAttrShm() failed")
		}

		log.Println("AttachAttrShm success!")
	})
}

type nodeUpdater func(pNode *node, attrID uint32, newVal uint32, oldVal *uint32) bool

func nodeSetVal(pNode *node, attrID uint32, newVal uint32, oldVal *uint32) bool {
	oldNodeVal := *pNode
	newNodeVal := makeNode(attrID, newVal)

	if oldNodeVal.ID() != attrID {
		return false
	}

	isSuccess := atomic.CompareAndSwapUint64(pNode.Uint64Ptr(), oldNodeVal.Uint64Val(), newNodeVal.Uint64Val())
	if nil != oldVal {
		*oldVal = oldNodeVal.Val()
	}

	return isSuccess
}

func nodeAddVal(pNode *node, attrID uint32, addition uint32, oldVal *uint32) bool {
	oldNodeVal := *pNode
	newNodeVal := makeNode(attrID, oldNodeVal.Val()+addition)

	if oldNodeVal.ID() != attrID {
		return false
	}

	isSuccess := atomic.CompareAndSwapUint64(pNode.Uint64Ptr(), oldNodeVal.Uint64Val(), newNodeVal.Uint64Val())
	if nil != oldVal {
		*oldVal = oldNodeVal.Val()
	}

	return isSuccess
}

var (
	try1Cnt [kMaxAttemp]uint64
	try2Cnt [kMaxAttemp]uint64
	try3Cnt [kMaxAttemp]uint64

	mu        sync.Mutex
	try2Attrs []uint32

	err1Cnt uint64
	err2Cnt uint64
	err3Cnt uint64
)

func ResetConflictsCounting() {
	for i := 0; i < kMaxAttemp; i++ {
		atomic.StoreUint64(&try1Cnt[i], 0)
		atomic.StoreUint64(&try2Cnt[i], 0)
		atomic.StoreUint64(&try3Cnt[i], 0)
	}

	atomic.StoreUint64(&err1Cnt, 0)
	atomic.StoreUint64(&err2Cnt, 0)
	atomic.StoreUint64(&err3Cnt, 0)
}

func ShowConflicts(file *os.File) {
	var totalCnt uint64
	for i := 0; i < kMaxAttemp; i++ {
		totalCnt += atomic.LoadUint64(&try1Cnt[i])
		totalCnt += atomic.LoadUint64(&try2Cnt[i])
		totalCnt += atomic.LoadUint64(&try3Cnt[i])
	}

	err1 := atomic.LoadUint64(&err1Cnt)
	err2 := atomic.LoadUint64(&err2Cnt)
	err3 := atomic.LoadUint64(&err3Cnt)

	totalCnt += err1
	totalCnt += err2
	totalCnt += err3

	fmt.Fprintf(file, "total cnt: %d\n", totalCnt)
	for i := 0; i < kMaxAttemp; i++ {
		if try1Cnt[i] != 0 {
			fmt.Fprintf(file, "attemp1[%2d]: %v, ration: %v%%\n", i, try1Cnt[i],
				100*float64(try1Cnt[i])/float64(totalCnt))
		}

		if try2Cnt[i] != 0 {
			fmt.Fprintf(file, "attemp2[%2d]: %v, ration: %v%%\n", i, try2Cnt[i],
				100*float64(try2Cnt[i])/float64(totalCnt))
		}

		if try3Cnt[i] != 0 {
			fmt.Fprintf(file, "attemp3[%2d]: %v, ration: %v%%\n", i, try3Cnt[i],
				100*float64(try3Cnt[i])/float64(totalCnt))
		}
	}

	if err1 != 0 {
		fmt.Fprintf(file, "err1 cnt: %d, ration: %v%%\n", err1, 100*float64(err1)/float64(totalCnt))
	}
	if err2 != 0 {
		fmt.Fprintf(file, "err2 cnt: %d, ration: %v%%\n", err2, 100*float64(err2)/float64(totalCnt))
	}
	if err3 != 0 {
		fmt.Fprintf(file, "err3 cnt: %d, ration: %v%%\n", err3, 100*float64(err3)/float64(totalCnt))
	}

	for i, v := range try2Attrs {
		fmt.Fprintf(file, "try2Attrs[%d] %d\n", i, v)
	}
}

func dumpShm(shm *uint8, prefix string) {
	pNode := (*[kNodeNum]uint64)(unsafe.Pointer(shm))
	for i := 0; i < kHashShmLen*kHashShmTimes; i++ {
		if pNode[i] != 0 {
			log.Printf("%s node[%4d] = (%s)\n", prefix, i, node(pNode[i]))
		}
	}
}

func DumpSumShm() {
	log.Printf("sum start addr:%p\n", sumShm)
	dumpShm(sumShm, "sum")
}

func DumpSetShm() {
	log.Printf("set start addr:%p\n", sumShm)
	dumpShm(setShm, "set")
}

// 根据Attr_API_New.c的逻辑重新实现一遍
func createOrUpdateNode(shm *uint8, attrID, newVal uint32, updater nodeUpdater, pOldVal *uint32, isSet bool) int {
	if 0 == attrID || nil == shm {
		atomic.AddUint64(&err2Cnt, 1)
		return -2
	}

	var (
		emptyNodes     [kHashShmTimes]*node
		zeroValNodes   [kHashShmTimes]*node
		emptyNodeCnt   int
		zeroValNodeCnt int
		pNode          *node

		hashVal   uint32
		pNodeAddr uintptr
		nodeCopy  node
	)

	for try := 0; try < kMaxAttemp; try++ {
		emptyNodeCnt = 0
		zeroValNodeCnt = 0

		for i := uint32(0); i < kHashShmTimes; i++ {
			hashVal = attrID % hashModes[i]
			pNodeAddr = (uintptr)(unsafe.Pointer(shm)) + (uintptr)(kHashShmLen*i+hashVal*kNodeSize)
			pNode = (*node)(unsafe.Pointer(pNodeAddr))

			if attrID == pNode.ID() { // 找到对应节点, 更新
				if updater(pNode, attrID, newVal, pOldVal) {
					atomic.AddUint64(&(try1Cnt[try]), 1)
					return 0
				}
			} else if 0 == pNode.ID() {
				emptyNodes[emptyNodeCnt] = pNode
				emptyNodeCnt += 1
			} else if 0 == pNode.Val() {
				zeroValNodes[zeroValNodeCnt] = pNode
				zeroValNodeCnt += 1
			}
		}

		newNode := makeNode(attrID, newVal)

		// 未找到节点，则做两次尝试
		// 1) 尝试在attrid==0的空闲节点上新建节点
		if emptyNodeCnt > 0 {
			for i := 0; i < emptyNodeCnt; i++ {
				pNode = emptyNodes[i]
				if 0 == pNode.ID() {
					if atomic.CompareAndSwapUint64(pNode.Uint64Ptr(), uint64(0), newNode.Uint64Val()) {
						if nil != pOldVal {
							*pOldVal = 0
						}

						atomic.AddUint64(&(try2Cnt[try]), 1)

						mu.Lock()
						try2Attrs = append(try2Attrs, attrID)
						mu.Unlock()

						return 0
					}
				}
			}
		}

		// 2) 尝试在value==0的空闲节点上新建节点
		if !isSet {
			for i := int32(zeroValNodeCnt - 1); i >= 0; i-- {
				pNode = zeroValNodes[i]
				nodeCopy = *pNode
				if 0 == nodeCopy.Val() {
					if atomic.CompareAndSwapUint64(pNode.Uint64Ptr(), nodeCopy.Uint64Val(), newNode.Uint64Val()) {
						if nil != pOldVal {
							*pOldVal = 0
						}

						atomic.AddUint64(&(try3Cnt[try]), 1)
						return 0 // 成功则返回
					}
				}
			}
		}
	}

	atomic.AddUint64(&err3Cnt, 1)
	return -3 // 所有尝试都失败
}

func AttrAPI(id, val uint32) int {
	return createOrUpdateNode(sumShm, id, val, nodeAddVal, nil, false)
}

func AttrAPISet(attrID, newVal uint32) int {
	shm := setShm
	if 0 == attrID || nil == shm {
		atomic.AddUint64(&err2Cnt, 1)
		return -2
	}

	var (
		emptyNodes   [kHashShmTimes]*node
		emptyNodeCnt int
		pNode        *node

		hashVal   uint32
		pNodeAddr uintptr
	)

	newNode := makeNode(attrID, newVal)
	for try := 0; try < kMaxAttemp; try++ {
		emptyNodeCnt = 0

		for i := uint32(0); i < kHashShmTimes; i++ {
			hashVal = attrID % hashModes[i]
			pNodeAddr = (uintptr)(unsafe.Pointer(shm)) + (uintptr)(kHashShmLen*i+hashVal*kNodeSize)
			pNode = (*node)(unsafe.Pointer(pNodeAddr))

			if attrID == pNode.ID() { // 找到对应节点, 更新
				if nodeSetVal(pNode, attrID, newVal, nil) {
					atomic.AddUint64(&(try1Cnt[try]), 1)
					return 0
				} else {
					goto next_loop
				}
			} else if 0 == pNode.ID() {
				emptyNodes[emptyNodeCnt] = pNode
				emptyNodeCnt += 1
			}
		}

		// 尝试在attrid==0的空闲节点上新建节点
		if emptyNodeCnt > 0 {
			for i := 0; i < emptyNodeCnt; i++ {
				pNode = emptyNodes[i]
				if 0 == pNode.ID() {
					if atomic.CompareAndSwapUint64(pNode.Uint64Ptr(), pNode.Uint64Val(), newNode.Uint64Val()) {
						atomic.AddUint64(&(try2Cnt[try]), 1)
						mu.Lock()
						try2Attrs = append(try2Attrs, attrID)
						mu.Unlock()
						return 0
					} else {
						goto next_loop
					}
				}
			}
		}
	next_loop:
	}

	atomic.AddUint64(&err3Cnt, 1)
	return -3 // 所有尝试都失败
}
