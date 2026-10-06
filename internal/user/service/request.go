package service

type UserCallback struct {
	Code  string `json:"code" form:"code" validate:"required"`
	State string `json:"state" form:"state" validate:"required"`
}

type UserUpdate struct {
	Nickname string `json:"nickname" form:"nickname" validate:"required && max:255"`
	Avatar   string `json:"avatar" form:"avatar" validate:"required && url && regex:\"^https?://\" && max:255"`
}

// LoginURL is the OAuth authorization URL the client should open.
type LoginURL struct {
	URL string `json:"url"`
}

// LoginToken is the bearer token for later requests.
type LoginToken struct {
	Token string `json:"token"`
}
