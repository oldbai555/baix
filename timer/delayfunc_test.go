/**
*  针对 delayFunc.go 做单元测试，主要测试延迟函数结构体是否正常使用
 */
package timer

import (
	"github.com/oldbai555/lbtool/log"
	"testing"
)

func SayHello(message ...interface{}) {
	log.Infof(message[0].(string), " ", message[1].(string))
}

func TestDelayfunc(t *testing.T) {
	df := NewDelayFunc(SayHello, []interface{}{"hello", "zinx!"})
	log.Infof("df.String() = %s", df.String())
	df.Call()
}
