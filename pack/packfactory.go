package pack

import (
	"github.com/oldbai555/baix/iface"
	"sync"
)

var packOnce sync.Once

type packFactory struct{}

var factoryInstance *packFactory

//生成不同封包解包的方式，单例

func Factory() *packFactory {
	packOnce.Do(func() {
		factoryInstance = new(packFactory)
	})

	return factoryInstance
}

// NewPack 创建一个具体的拆包解包对象
func (f *packFactory) NewPack(kind string) iface.IDataPack {
	var dataPack iface.IDataPack

	switch kind {
	// BaiX 标准默认封包拆包方式
	case iface.BaiXDataPack:
		dataPack = newDataPack()
		break
	default:
		dataPack = newDataPack()
	}

	return dataPack
}
