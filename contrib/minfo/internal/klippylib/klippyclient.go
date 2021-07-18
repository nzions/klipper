package klippylib

import (
	"encoding/json"
	"fmt"
	"net"
)

func NewClient(path string) *Client {
	return &Client{
		SockPath: path,
	}
}

type Client struct {
	SockPath string

	c net.Conn
}

func (x *Client) Dial() (c *Client, err error) {
	x.c, err = net.Dial("unix", x.SockPath)
	if err != nil {
		return nil, err
	}
	return x, nil
}

func (x *Client) Close() {
	x.c.Close()
}

func (x *Client) doCmd(c Command, resp *Response) error {
	jBytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	// // add the trminator
	jBytes = append(jBytes, 0x03)
	_, err = x.c.Write(jBytes)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	n, err := x.c.Read(buf[:])
	if err != nil {
		return err
	}

	// dirty hack
	dBytes := buf[0 : n-1]
	// fmt.Println(string(dBytes))
	err = json.Unmarshal(dBytes, resp)
	if err != nil {
		fmt.Println("ERR:", err, string(dBytes))
		return err
	}
	return nil
}

func (x *Client) GetInfo() (resp InfoResponse, err error) {
	cmd := Command{
		ID:     1,
		Method: "info",
	}

	r := &Response{}
	r.Result = &resp
	err = x.doCmd(cmd, r)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// objects/list
// info
// "objects/subscribe", "params":
// {"objects":{"toolhead": ["position"], "webhooks": ["state"]},
// "response_template":{}}}`
// func (x *Client) GetInfo() {
// 	fmt.Println("getting info")
// 	cmd := Command{
// 		ID:     123,
// 		Method: "objects/list",
// 	}
// 	x.doCmd(cmd, nil)
// }

// // objects/list
// info

func (x *Client) GetMCUInfo() (resp MCUResponse, err error) {
	objs := CommandObjectList{}
	objs.Objects = make(map[string][]string)
	objs.Objects["mcu"] = []string{"mcu_version", "mcu_build_versions", "mcu_constants"}

	cmd := Command{
		ID:     1,
		Method: "objects/query",
		Params: &objs,
	}

	r := &Response{}
	r.Result = &resp
	return resp, x.doCmd(cmd, r)
}
