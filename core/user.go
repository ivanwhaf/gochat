package core

type Sex int32

const (
	SexMale   Sex = 0
	SexFemale Sex = 1
)

type Auth int32

const (
	AuthAdmin Auth = 0
	AuthUser  Auth = 1
)

type User struct {
	Id       int32
	Nickname string
	Sex      Sex
	Password string
	Auth     Auth
}
