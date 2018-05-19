package domain

type SkillContract struct {
	Producer string `form:"producer"`
	Consumer string `form:"consumer"`
	Content  string `form:"content"`
	Platform string `form:"platform"`
	Price    uint32 `form:"price"`
	Ratio    uint8  `form:"ratio"`
}
