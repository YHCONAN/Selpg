package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Selpg struct { //定义selpg所有参数的结构体
	start_page  int    //开始页号
	end_page    int    //结束页号
	in_filename string //输入文件名
	page_len    int    //页面长度
	page_type   bool   //分割页面的方式
	print_dest  string //目标打印机
}

func get_args(selpg *Selpg) { //获取参数
	if selpg == nil { //selpg为空
		fmt.Fprintf(os.Stderr, "\nSlepg can not be nil\n\n")
		os.Exit(1)
	}

	//绑定参数
	flag.IntVar(&selpg.start_page, "s", -1, "Start Page")
	flag.IntVar(&selpg.end_page, "e", -1, "End Page")
	flag.IntVar(&selpg.page_len, "l", 7, "Lines Per Page") //便于测试
	//flag.IntVar(&selpg.page_len, "l", 72, "Lines Per Page")
	flag.BoolVar(&selpg.page_type, "f", false, "Delimited By Form Feeds")
	flag.StringVar(&selpg.print_dest, "d", "", "Specify The Printer")
	flag.Parse()

	//处理non-flag参数，若存在则为输入文件名
	if len(flag.Args()) > 0 {
		selpg.in_filename = flag.Args()[0]
	} else {
		selpg.in_filename = ""
	}
}

func process_args(selpg *Selpg) { //处理参数
	get_args(selpg) //获取参数

	if selpg.start_page == -1 || selpg.end_page == -1 { //没有输入开始页和结束页
		fmt.Fprintf(os.Stderr, "\nStart_Page and End_Page are needed\n\n")
		os.Exit(2)
	} else if selpg.start_page <= 0 || selpg.end_page <= 0 || selpg.start_page > selpg.end_page || selpg.page_len <= 0 { //参数不合法
		fmt.Fprintf(os.Stderr, "\nStart_Page or End_Page or Page_Len are invalid\n\n")
		os.Exit(3)
	}
}

func process_input(selpg *Selpg) { //输入输出
	//确定输入文件
	var fin *os.File

	if selpg.in_filename == "" { //文件名为空
		fin = os.Stdin
	} else { //文件名非空
		var err error
		fin, err = os.Open(selpg.in_filename)

		if err != nil {
			fmt.Fprintf(os.Stderr, "\ncan not open file %s\n\n", selpg.in_filename)
			os.Exit(4)
		}
	}

	//确定输出文件
	var fout io.WriteCloser

	if selpg.print_dest == "" { //打印目的地为空
		fout = os.Stdout
	} else { //打印目的地非空
		cmd := exec.Command("lp -d " + selpg.print_dest)
		//cmd := exec.Command("cat") //可用于测试
		var err error
		fout, err = cmd.StdinPipe() //将当前的输出作为cmd的输入
		if err != nil {
			fmt.Fprintf(os.Stderr, "\ncan not open pipe\n\n")
			os.Exit(5)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start() //开始执行
		if err != nil {
			fmt.Fprintf(cmd.Stderr, "\nlp: The printer or class does not exist\n\n")
			os.Exit(5)
		}
	}

	if !selpg.page_type { //按照page_len确定页
		line_ctr := 0
		page_ctr := 1
		buf := bufio.NewReader(fin)

		for true {
			line, err := buf.ReadString('\n')
			if err != nil {
				if err == io.EOF { //读至文件末尾
					break
				}
				fmt.Fprintf(os.Stderr, "\nread file error\n\n")
				os.Exit(6)
			}

			line_ctr++
			if line_ctr > selpg.page_len { //读够一页
				page_ctr++
				line_ctr = 1
			}
			if (page_ctr >= selpg.start_page) && (page_ctr <= selpg.end_page) { //写入
				_, err := fout.Write([]byte(line))
				if err != nil {
					fmt.Fprintf(os.Stderr, "\nwrite file error\n\n")
					os.Exit(7)
				}
			} else if page_ctr > selpg.end_page {
				break
			}
		}
	} else { //按照换页符确定页
		page_ctr := 0
		buf := bufio.NewReader(fin)

		for true {
			//page, err := buf.ReadString('\f')
			page, err := buf.ReadString('\n') //便于测试
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintf(os.Stderr, "\nread file error\n\n")
				os.Exit(6)
			}
			page_ctr++

			if (page_ctr >= selpg.start_page) && (page_ctr <= selpg.end_page) { //写入
				_, err := fout.Write([]byte(page))
				if err != nil {
					fmt.Fprintf(os.Stderr, "\nwrite file error\n\n")
					os.Exit(7)
				}
			} else if page_ctr > selpg.end_page {
				break
			}
		}
	}

}

func main() {
	var selpg Selpg       //创建结构体
	process_args(&selpg)  //处理参数
	process_input(&selpg) //输入输出
}
