package utils

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type UserInfo struct {
	Ret         int
	Id          int
	AccountName string
	LoginToken  string
	Isadmin     int
	ResultMsg   string
	Pwd         string
}

var (
	dbhostsip  = "127.0.0.1:3306" //IP地址
	dbusername = "root"           //用户名
	dbpassword = "123456"         //密码
	dbname     = "gloudfinance"   //库名
)

var db *sql.DB

func init() {
	fmt.Print("connect...............")
	mysqldb, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", dbusername, dbpassword, dbhostsip, dbname))

	if err != nil {
		fmt.Print("connect DB error...")
	}

	db = mysqldb
	//defer db.Close()
}

/**************************

		用户相关

***************************/

//生成随机字符串
func GetRandomString(strLen int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < strLen; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func GetLoginToken() string {
	return GetRandomString(32)
}

func GetMd5Str(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str1
}

func QueryUserInfo(account, pwd, logintoken string) UserInfo {
	// 保存用户信息列表
	fmt.Println(" GetMd5Str(pwd)", GetMd5Str(pwd))
	var userinfo UserInfo
	if nil == db {
		fmt.Print("DB IS NULL.......")
		userinfo.Ret = -1
		userinfo.ResultMsg = "数据库初始化异常"
	} else {
		var querySql string
		if len(pwd) > 0 {
			querySql = fmt.Sprintf("SELECT * FROM account where accountName='%s' and pwd='%s'", account, GetMd5Str(pwd))
		} else {
			querySql = fmt.Sprintf("SELECT * FROM account where accountName='%s' and logintoken='%s'", account, logintoken)
		}
		rows, err := db.Query(querySql)
		if err != nil {
			//有异常
			userinfo.Ret = -2
			userinfo.ResultMsg = "查询异常"
			fmt.Print("QueryUserinfo error1....", err.Error())
		}

		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&userinfo.Id, &userinfo.AccountName, &userinfo.Pwd, &userinfo.LoginToken, &userinfo.Isadmin)
			if err != nil {
				userinfo.Ret = -3
				userinfo.ResultMsg = "查询失败"
				fmt.Print("QueryUserinfo error", err.Error())
			}
		}
		err = rows.Err()
		if err != nil {
			userinfo.Ret = -4
			userinfo.ResultMsg = "查询失败，未找到当前用户信息"
			fmt.Print("QueryUserinfo error2...")
		}

		if userinfo.Id <= 0 {
			userinfo.Ret = -5
			if len(logintoken) > 0 {
				userinfo.ResultMsg = "登录令牌错误"
			} else {
				userinfo.ResultMsg = "用户名或密码错误"
			}

			userinfo.ResultMsg = "用户名或密码错误"
		}
	}
	return userinfo
}

func UpdateUserToken(userid int) string {
	//更新token
	newToken := GetLoginToken()
	_, error := db.Exec(fmt.Sprintf("UPDATE account SET logintoken='%s' WHERE id=%d", newToken, userid))
	if error == nil {
		return newToken
	}
	return ""
}

func ReSetPwd(account, oldpwd, newpad, logintoken string) UserInfo {
	// 保存用户信息列表
	fmt.Println(" GetMd5Str(pwd)", GetMd5Str(oldpwd))
	var userinfo UserInfo
	if nil == db {
		fmt.Print("DB IS NULL.......")
		userinfo.Ret = -1
		userinfo.ResultMsg = "数据库初始化异常"
	} else {

		userinfo = QueryUserInfo(account, "", logintoken)
		if userinfo.Ret != 0 {
			//查询当前用户失败
			userinfo.ResultMsg = "当前用户身份验证失败"
		} else {
			if GetMd5Str(oldpwd) == userinfo.Pwd {
				//查询当前用户成功
				notoken := GetLoginToken()
				updateSql := fmt.Sprintf("UPDATE account SET logintoken='%s',pwd='%s' WHERE id=%d", notoken, GetMd5Str(newpad), userinfo.Id)

				fmt.Print("updateSql", updateSql)
				//更新token
				_, error := db.Exec(updateSql)
				if error == nil {
					userinfo.Ret = 0
					userinfo.ResultMsg = "重设密码成功"
					userinfo.LoginToken = notoken
				} else {
					userinfo.Ret = -1
					userinfo.ResultMsg = "重设密码失败"
				}
			} else {
				userinfo.Ret = -2
				userinfo.ResultMsg = "原始密码错误"
			}
		}

	}

	return userinfo
}

/**************************

		发票相关

***************************/

//未知,未知,发票左上角的号，发票号，金额，开票日期，校验码，未知
type InvoiceObj struct {
	InVoiceId   int    `json:"invoiceid"`
	LeftCode    string `json:"leftCode"`
	InvoiceCode string `json:"invoiceCode"`
	Price       string `json:"price"`
	Date        string `json:"date"`
	CheckCode   string `json:"checkCode"`
	OrignStr    string `json:"orignStr"`
	InfoMaster  string `json:"master"`
	Updater     string `json:"updater"`
	Update_Date int64  `json:"update_date"`
	Claim       int    `json:"claim"`
	UpdaterBy   string `json:"update_by"`
	IsDelete    int    `json:"isDelete"`
}

type InsertVoiceRespon struct {
	Ret  int        `json:"ret"`
	Msg  string     `json:"msg"`
	Data InvoiceObj `json:"data"`
}

func QueryInvoice(invoiceStr, account string) InsertVoiceRespon {
	var invoiceRespon InsertVoiceRespon

	//未知,未知,发票左上角的号，发票号，金额，开票日期，校验码，未知
	//01,04,1100173320,39984993,548.11,20180506,10842592691815483377,99B9,
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM invoice_table where orign_invoice_str='%s'", invoiceStr))
	if err != nil {
		//有异常
		invoiceRespon.Ret = -1
		invoiceRespon.Msg = "查询异常"
		fmt.Print("QueryUserinfo error1....", err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&invoiceRespon.Data.InVoiceId, &invoiceRespon.Data.OrignStr, &invoiceRespon.Data.InfoMaster, &invoiceRespon.Data.Updater, &invoiceRespon.Data.Update_Date, &invoiceRespon.Data.Claim, &invoiceRespon.Data.IsDelete, &invoiceRespon.Data.UpdaterBy)
		fmt.Println("invoiceRespon.Data.isDelete:", invoiceRespon.Data.IsDelete)
		if err != nil {

			invoiceRespon.Ret = -2
			invoiceRespon.Msg = "失败"
			fmt.Print("QueryUserinfo error", err.Error())
		}
	}
	err = rows.Err()
	if err != nil {
		invoiceRespon.Ret = -3
		invoiceRespon.Msg = "查询失败，未找到发票信息"
		fmt.Print("QueryUserinfo error2...")
	}

	//	t := time.Now()
	//	t.Unix()
	return invoiceRespon
}

func InsertVoice(invoiceStr, account, updater string) InsertVoiceRespon {

	var invoiceRespon InsertVoiceRespon
	invoiceRespon = QueryInvoice(invoiceStr, account)
	if invoiceRespon.Data.InVoiceId > 0 {
		//之前有数据，直接反回
		fmt.Println("之前有数据，直接反回", invoiceRespon.Data.IsDelete)
		invoiceRespon.Ret = 0
		invoiceRespon.Msg = "查询成功"
		return invoiceRespon
	} else {
		//没有数据，插入
		fmt.Println("没有数据，插入", invoiceStr)
		t := time.Now()
		fmt.Println("time..", invoiceStr, account, t.Unix())
		db.Exec("INSERT INTO invoice_table(orign_invoice_str,info_master,updater,update_date) values(?,?,?,?)", invoiceStr, account, updater, t.Unix())
	}
	return QueryInvoice(invoiceStr, account)
	//db.Exec("INSERT INTO user(uid,username,age) values(?,?,?)", i, "user"+strconv.Itoa(i), i-1000)
}

//更新发票是否已报销使用过
func UpdateInvoiceClaim(claim, invoiceid int, updaterby string) (int, string) {
	_, error := db.Exec(fmt.Sprintf("UPDATE invoice_table SET claim='%d',updater_by='%s' WHERE id='%d'", claim, updaterby, invoiceid))
	if error == nil {
		return 0, "更新成功"
	}
	return -1, "更新失败"
}

//更新表中删除标识
func DelInvoicem(isDelete, invoiceid int, updaterby string) (int, string) {
	_, error := db.Exec(fmt.Sprintf("UPDATE invoice_table SET isdelete='%d',updater_by='%s' WHERE id='%d'", isDelete, updaterby, invoiceid))
	if error == nil {
		return 0, "删除成功"
	}
	return -1, error.Error()
}
