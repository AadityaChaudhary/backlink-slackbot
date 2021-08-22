package main

import "fmt"

func main() {
	if err := InitDB(false); err != nil {
		fmt.Println(err)
		panic(err)
	}

	//DropAllTables()

	//if err := AddWorkspace("nag"); err != nil {
	//	fmt.Println(err)
	//	panic(err)
	//}

	//err := AddBacklinkToWorkspace("nag", Backlink{LinkName: "bushan", NotionID: "nagabushan"})
	//if err != nil {
	//	fmt.Println(err)
	//	panic(err)
	//}

	workspace := GetWorkspaceInfo("nag")
	fmt.Println(workspace)

	defer func() {
		if err := DeinitDB(); err != nil {
			fmt.Println(err)
		}
	}()
}
