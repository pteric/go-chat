package process

import (
	"encoding/json"
	"fmt"
	commen "go-chat/commen/message"
	"go-chat/server/model"
	"go-chat/server/utils"
	"net"
)

type UserProcess struct {
	Conn net.Conn
}

func register(userName, passWord, passWordConfirm string) (user model.User, err error) {
	user, err = model.CurrentUserDao.Register(userName, passWord, passWordConfirm)
	return
}

func login(userName, passWord string) (user model.User, err error) {
	// 判断用户名和密码
	user, err = model.CurrentUserDao.Login(userName, passWord)
	return
}

// 响应客户端
func (this *UserProcess) responseClient(responseMessageType string, code int, data string, err error) {
	var responseMessage commen.ResponseMessage
	responseMessage.Code = code
	responseMessage.Type = responseMessageType
	responseMessage.Data = data

	responseData, err := json.Marshal(responseMessage)
	if err != nil {
		fmt.Printf("some error when generate response message, error: %v", err)
	}

	dispatcher := utils.Dispatcher{Conn: this.Conn}

	err = dispatcher.WirteData(responseData)
}

func (this *UserProcess) UserRegister(message string) (err error) {
	var info commen.RegisterMessage
	var code int
	data := ""
	err = json.Unmarshal([]byte(message), &info)
	if err != nil {
		code = commen.ServerError
	}

	_, err = register(info.UserName, info.Password, info.PasswordConfirm)
	switch err {
	case nil:
		code = commen.RegisterSucceed
	case model.ERROR_PASSWORD_NOT_MATCH:
		code = 402
	case model.ERROR_USER_EXISTED:
		code = 403
	default:
		code = 500
	}
	this.responseClient(commen.RegisterResponseMessageType, code, data, err)
	return
}

func (this *UserProcess) UserLogin(message string) (err error) {
	var info commen.LoginMessage
	var code int
	var data string
	err = json.Unmarshal([]byte(message), &info)
	if err != nil {
		code = commen.ServerError
	}

	user, err := login(info.UserName, info.Password)

	switch err {
	case nil:
		code = commen.LoginSucceed
		// save user conn status
		clientConn := model.ClientConn{}
		clientConn.Save(user.ID, user.Name, this.Conn)

		userInfo := commen.UserInfo{user.ID, user.Name}
		info, _ := json.Marshal(userInfo)
		data = string(info)
	case model.ERROR_USER_NOT_EXISTS:
		code = 404
	case model.ERROR_USER_PWD:
		code = 403
	default:
		code = 500
	}
	this.responseClient(commen.LoginResponseMessageType, code, data, err)
	return
}
