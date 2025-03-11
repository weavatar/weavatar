package request

type UserCallback struct {
	Code  string `query:"code" form:"code" validate:"required"`
	State string `query:"state" form:"state" validate:"required"`
}

type UserUpdate struct {
	Nickname string `form:"nickname" json:"nickname" validate:"required"`
	Avatar   string `form:"avatar" json:"avatar" validate:"required|isFullURL"`
}
