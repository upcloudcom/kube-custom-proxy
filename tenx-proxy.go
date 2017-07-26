/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-07-20  @author yangle
 */

package main

import (
	"tenx-proxy/config/cmd"
)

func main() {
	cmds := cmd.NewCommand()
	cmds.Execute()
}
