package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"./utils"

	"./constantValues"
)

type UserInfoRespon struct {
	Id          int    `json:"account_id"`
	AccountName string `json:"account_name"`
	LoginToken  string `json:"logintoken"`
	Isadmin     int    `json:"is_admin"`
}

type LoginRespon struct {
	Ret  int            `json:"ret"`
	Msg  string         `json:"msg"`
	Data UserInfoRespon `json:"data"`
}

type BaseRespon struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}

//登录
func LoginHttpRequest(w http.ResponseWriter, req *http.Request) {
	utils.FileLogs.Trace("URL:%s>>IP: %s>>UA:%s", req.RequestURI, req.RemoteAddr, req.UserAgent())
	if req.Method == "GET" {
		req.ParseForm()
		t := req.Form.Get(constantValues.T)
		if t == constantValues.LOGIN {
			//登录方法
			account := req.Form.Get(constantValues.ACCOUNT)
			pwd := req.Form.Get(constantValues.PWD)
			logintoken := req.Form.Get(constantValues.LOGINTOEKN)
			userinfo := utils.QueryUserInfo(account, pwd, logintoken)

			var loginrespson LoginRespon
			loginrespson.Ret = userinfo.Ret
			loginrespson.Msg = userinfo.ResultMsg

			if userinfo.Ret == 0 {
				//登录成功，更新token
				loginrespson.Data.Id = userinfo.Id
				loginrespson.Data.AccountName = userinfo.AccountName
				loginrespson.Data.Isadmin = userinfo.Isadmin
				loginrespson.Data.LoginToken = utils.UpdateUserToken(userinfo.Id)
				loginrespson.Msg = "登录成功"
			}

			result, _ := json.Marshal(loginrespson)
			resultjson := string(result)
			io.WriteString(w, resultjson)
			utils.FileLogs.Trace("URL:%s>>IP: %s>>UA:%s", req.RequestURI, req.RemoteAddr, req.UserAgent(), " Result=", resultjson)
		} else if t == constantValues.RESETPWD {
			//重设密码
			fmt.Print("t == constantValues.RESETPWD")
			account := req.Form.Get(constantValues.ACCOUNT)
			oldpwd := req.Form.Get(constantValues.OLD_PWD)
			newpwd := req.Form.Get(constantValues.NEW_PWD)
			logintoken := req.Form.Get(constantValues.LOGINTOEKN)
			userinfo := utils.ReSetPwd(account, oldpwd, newpwd, logintoken)

			var loginrespson LoginRespon
			loginrespson.Ret = userinfo.Ret
			loginrespson.Msg = userinfo.ResultMsg
			loginrespson.Data.LoginToken = userinfo.LoginToken
			loginrespson.Data.Id = userinfo.Id
			loginrespson.Data.AccountName = userinfo.AccountName
			loginrespson.Data.Isadmin = userinfo.Isadmin

			result, _ := json.Marshal(loginrespson)
			resultjson := string(result)
			io.WriteString(w, resultjson)
			utils.FileLogs.Trace("URL:%s>>IP: %s>>UA:%s", req.RequestURI, req.RemoteAddr, req.UserAgent(), " Result=", resultjson)
		}

	} else if req.Method == "POST" {
		account := req.PostFormValue("account")
		b := req.Form.Get("pwd")
		fmt.Print("account==" + account)
		fmt.Print("pwd==" + b)
		result, _ := ioutil.ReadAll(req.Body)
		req.Body.Close()
		fmt.Printf("%s\n", result)

	}

}

//发票信息处理相关
//未知,未知,发票左上角的号，发票号，金额，开票日期，校验码，未知
//01,04,1100173320,39984993,548.11,20180506,10842592691815483377,99B9,
func InvoiceStatusHttpRequest(w http.ResponseWriter, req *http.Request) {
	utils.FileLogs.Trace("URL:%s>>IP: %s>>UA:%s", req.RequestURI, req.RemoteAddr, req.UserAgent())
	if req.Method == "GET" {
		req.ParseForm()
		t := req.Form.Get(constantValues.T)
		if t == constantValues.INVOICE_INSERT {
			//插入数据
			invoiceStr := req.Form.Get(constantValues.INVOICE_DATA)
			account := req.Form.Get(constantValues.ACCOUNT)
			logintoken := req.Form.Get(constantValues.LOGINTOEKN)
			currentAccount := req.Form.Get(constantValues.UPDATER)
			userinfo := utils.QueryUserInfo(currentAccount, "", logintoken)

			var invoiceRespon utils.InsertVoiceRespon
			if len(invoiceStr) <= 0 || len(account) <= 0 || len(logintoken) <= 0 || len(currentAccount) <= 0 {
				//发票信息没传
				invoiceRespon.Ret = -10
				invoiceRespon.Msg = "缺少必要参数"
			} else {
				if userinfo.Ret == 0 {
					//用户验证通过
					invoiceRespon = utils.InsertVoice(invoiceStr, account, currentAccount)
					invoiceData := strings.Split(invoiceRespon.Data.OrignStr, ",")
					fmt.Println("str==", invoiceRespon.Data.OrignStr)
					if len(invoiceData) > 6 {
						//合法数据不会数组越界
						//未知,未知,发票左上角的号，发票号，金额，开票日期，校验码，未知
						//						var invoiceInfo InsertVoice
						invoiceRespon.Data.LeftCode = invoiceData[2]
						invoiceRespon.Data.InvoiceCode = invoiceData[3]
						invoiceRespon.Data.Price = invoiceData[4]
						invoiceRespon.Data.Date = invoiceData[5]
						invoiceRespon.Data.CheckCode = invoiceData[6]

						//invoiceInfoStr, _ := json.Marshal(invoiceInfo)
						if len(invoiceRespon.Msg) <= 0 {
							invoiceRespon.Msg = "数据检验成功"
						}

						//						fmt.Print("invoice=", invoiceRespon.Data)
					} else {
						invoiceRespon.Ret = -11
						invoiceRespon.Msg = "未知的发票"
					}
				} else {
					invoiceRespon.Ret = userinfo.Ret
					invoiceRespon.Msg = "验证用户信息失败"
				}

			}
			result, _ := json.Marshal(invoiceRespon)
			resultjson := string(result)
			io.WriteString(w, resultjson)
			utils.FileLogs.Trace("URL:%s>>IP: %s>>UA:%s", req.RequestURI, req.RemoteAddr, req.UserAgent(), " Result=", resultjson)
		} else if t == constantValues.INVOICE_UPDATE {
			//更新数据
			var baseRespon BaseRespon
			account := req.Form.Get(constantValues.ACCOUNT)
			invoiceid := req.Form.Get(constantValues.INVOICE_ID)
			logintoken := req.Form.Get(constantValues.LOGINTOEKN)
			isClaim := req.Form.Get(constantValues.INVOICE_CLAIM)

			if len(isClaim) <= 0 || len(account) <= 0 || len(logintoken) <= 0 || len(invoiceid) <= 0 {
				//发票信息没传
				baseRespon.Ret = -10
				baseRespon.Msg = "缺少必要参数"
			} else {

				claimInt, err := strconv.Atoi(isClaim)
				invoiceIdInt, err1 := strconv.Atoi(invoiceid)

				if claimInt >= 0 && invoiceIdInt > 0 && err == nil && err1 == nil {
					userinfo := utils.QueryUserInfo(account, "", logintoken)

					baseRespon.Ret = userinfo.Ret
					baseRespon.Msg = userinfo.ResultMsg

					if userinfo.Ret == 0 {
						baseRespon.Ret, baseRespon.Msg = utils.UpdateInvoiceClaim(claimInt, invoiceIdInt, account)
					} else {
						baseRespon.Ret = userinfo.Ret
						baseRespon.Msg = userinfo.ResultMsg

					}
				} else {
					baseRespon.Ret = -10
					baseRespon.Msg = "参数错误"
				}

			}

			result, _ := json.Marshal(baseRespon)
			resultjson := string(result)
			io.WriteString(w, resultjson)
			utils.FileLogs.Trace("URL:%s>>IP: %s>>UA:%s", req.RequestURI, req.RemoteAddr, req.UserAgent(), " Result=", resultjson)
		} else if t == constantValues.INVOICE_DEL {
			//删除数据
			var baseRespon BaseRespon
			account := req.Form.Get(constantValues.ACCOUNT)
			invoiceid := req.Form.Get(constantValues.INVOICE_ID)
			logintoken := req.Form.Get(constantValues.LOGINTOEKN)
			isdelete := req.Form.Get(constantValues.INVOICE_IS_DEL)
			if len(account) <= 0 || len(logintoken) <= 0 || len(invoiceid) <= 0 {
				//发票信息没传
				baseRespon.Ret = -10
				baseRespon.Msg = "缺少必要参数"
			} else {

				invoiceIdInt, err1 := strconv.Atoi(invoiceid)

				if invoiceIdInt > 0 && err1 == nil {
					userinfo := utils.QueryUserInfo(account, "", logintoken)

					baseRespon.Ret = userinfo.Ret
					baseRespon.Msg = userinfo.ResultMsg

					if userinfo.Ret == 0 {
						IsDelete, _ := strconv.Atoi(isdelete)

						if IsDelete == 1 && userinfo.Isadmin == 1 {
							//1代表超管，只有超管可删除
							baseRespon.Ret, baseRespon.Msg = utils.DelInvoicem(IsDelete, invoiceIdInt, account)
						} else if IsDelete == 0 {
							baseRespon.Ret, baseRespon.Msg = utils.DelInvoicem(IsDelete, invoiceIdInt, account)
						} else {
							baseRespon.Ret = -12
							baseRespon.Msg = "无权操作"
						}

					} else {
						baseRespon.Ret = userinfo.Ret
						baseRespon.Msg = userinfo.ResultMsg

					}
				} else {
					baseRespon.Ret = -11
					baseRespon.Msg = "参数错误"
				}

			}

			result, _ := json.Marshal(baseRespon)
			resultjson := string(result)
			io.WriteString(w, resultjson)
			utils.FileLogs.Trace("URL:%s>>IP: %s>>UA:%s", req.RequestURI, req.RemoteAddr, req.UserAgent(), " Result=", resultjson)
		}

	} else if req.Method == "POST" {

	}
}

func main() {
	http.HandleFunc("/User", LoginHttpRequest)
	http.HandleFunc("/Invoice", InvoiceStatusHttpRequest)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
