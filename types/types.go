package types

type Channel struct {
	ID   string `json:"id"`
	Href string `json:"href"`
	Img  string `json:"img"`
}

type Config struct {
	Base        string `yaml:"base"`
	DefaultLogo string `yaml:"default_logo"`
	Port        string `yaml:"port"`
}
