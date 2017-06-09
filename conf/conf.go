package conf

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/snowflake"
)

var (
	SF *snowflake.Node
)

var Exports []ExportsItem

func init() {

	LoadSvrConf()

	LoadExports()
}

type ExportsItem struct {
	ExportPath string
	Info       []ExportsInfo
}

type ExportsInfo struct {
	GroupName  string
	Hosts      []string
	Permission string
	Squash     string
	WriteMode  string
}

type NfsproItem struct {
	Com string `json:"com"`
}

var AppConf struct {
	Host        string        `json:"Host"`
	PmapHost    string        `json:"PmapHost"`
	MountPort   uint32        `json:"MountPort"`
	NfsPort     uint32        `json:"NfsPort"`
	NfsaclPort  uint32        `json:"NfsaclPort"`
	Exports     []ExportsItem `json:"Exports"`
	Snowflakeid int64         `json:"snowflakeid"`
	LogLvl      int           `json:"loglvl"`
}

func AnalysisIp(str string) []string {

	var ips []string

	point := strings.Split(str, ".")

	tmp_index := len(point)

	if tmp_index < 4 {
		fmt.Println("conf AnalysisIp error")
		os.Exit(1)
	}

	last_point := point[tmp_index-1]

	if strings.Contains(last_point, "/") == true {

		indexs := strings.Split(last_point, "/")

		if len(indexs) != 2 {
			fmt.Println("conf AnalysisIp error")
			os.Exit(1)
		}

		begin, err := strconv.Atoi(indexs[0])

		if err != nil {
			fmt.Println("conf AnalysisIp atoi error:", err)
			os.Exit(1)
		}

		end, err := strconv.Atoi(indexs[1])

		if err != nil {
			fmt.Println("conf AnalysisIp atoi error:", err)
			os.Exit(1)
		}

		for i := begin; i < end; i++ {
			front_point := point[0:3]
			front_ip := strings.Join(front_point, ".")
			ip := front_ip + "." + strconv.Itoa(i)
			ips = append(ips, ip)
		}

	} else {
		ips = append(ips, str)
	}

	return ips
}

func LoadSvrConf() {

	conf_data, err := ioutil.ReadFile("./conf/svr.json")

	if err != nil {
		fmt.Println("conf ReadFile error:", err)
		os.Exit(0)
	}

	json_err := json.Unmarshal(conf_data, &AppConf)

	if json_err != nil {
		fmt.Println("conf json Unmarshal error :", json_err)
		os.Exit(1)
	}

	SF, err = snowflake.NewNode(AppConf.Snowflakeid)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func LoadExports() {

	f, err := os.OpenFile("/etc/exports", os.O_RDONLY, os.ModePerm)

	if err != nil {
		fmt.Println("LoadExports open /etc/exports error :", err)
		os.Exit(1)
	}

	defer f.Close()

	r := bufio.NewReader(f)

	for {

		b, _, err := r.ReadLine()

		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		item_str := strings.TrimLeft(string(b), " ")

		if len(item_str) <= 0 || item_str[0] == '#' || item_str == "" {
			continue
		}

		//fmt.Println("LoadExports item str :", item_str)

		path_info := strings.SplitN(item_str, " ", 2)

		if len(path_info) < 2 {
			fmt.Println("LoadExports item path info error")
			os.Exit(1)
		}

		path := path_info[0]
		ips_infos_all_str := path_info[1]

		fmt.Println("LoadExports item path :", path, "ips_infos_all_str :", ips_infos_all_str)

		ips_infos := strings.Split(ips_infos_all_str, " ")

		var item ExportsItem

		item.ExportPath = path

		for _, ips_infos_str := range ips_infos {

			index := strings.Index(ips_infos_str, "(")

			ips_str := ips_infos_str[0:index]
			params := strings.Split(strings.TrimRight(ips_infos_str[index:], ")"), ",")

			var exports_info ExportsInfo

			ips := AnalysisIp(ips_str)

			fmt.Println("LoadExports ips :", ips)

			exports_info.Hosts = ips
			exports_info.GroupName = ips_str

			for _, p := range params {

				if p == "ro" || p == "rw" {
					exports_info.Permission = p
				}

				if p == "root_squash" || p == "no_root_squash" || p == "all_squash" {
					exports_info.Squash = p
				}

				if p == "sync" || p == "async" {
					exports_info.WriteMode = p
				}
			}

			item.Info = append(item.Info, exports_info)

		}

		//fmt.Printf("item :%+v\n", exports_info)

		Exports = append(Exports, item)
	}

	for _, k := range Exports {
		fmt.Printf("%+v\n", k)
	}
}

func IsExport(path string, ip string) bool {

	var flag bool

	for _, i := range Exports {
		if i.ExportPath == path {
			for _, h := range i.Info {
				for _, k := range h.Hosts {
					if ip == k {
						flag = true
						break
					}
				}
			}
		}
	}

	return flag
}
