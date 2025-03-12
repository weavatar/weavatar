package biz

type QqHash struct {
	H []byte `gorm:"primaryKey"`
	Q string
}
