// 冒烟脚本用：按私钥加密明文密码，输出 hex 密文。
package main

import (
	"fmt"
	"os"

	"github.com/jiangbyte/light-gozero/common/pwdcipher"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: pwdenc <private_key_hex> <plain>")
		os.Exit(2)
	}
	c, err := pwdcipher.NewFromPrivateKeyHex(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out, err := c.Encrypt(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(out)
}
