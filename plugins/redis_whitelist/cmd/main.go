package main

type Options struct {
	RedisIDs        string `short:"s" long:"rdsID" required:"true"`
	RegionId        string `short:"r" long:"regionID" required:"true"`
	WhiteListName   string `short:"d" long:"Uniq whitelist name for this pod" required:"true"`
	AccessKeyID     string
	AccessKeySecret string
	ToDelete        bool `long:"delete"`
}

var (
	opt    Options
	rdsIDs []string
)

func init() {

}

func main() {

}
