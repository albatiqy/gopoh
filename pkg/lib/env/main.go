package env

import (
	//"os"
	//"time"

	"github.com/albatiqy/gopoh/contract/log"
	"github.com/albatiqy/gopoh/internal/support/subosito/gotenv"
)

/*
var (
	TimeZone string
	Loc      *time.Location
)
*/

func Load(files ...string) {
	if err := gotenv.Load(files...); err != nil {
		log.Fatal(`tidak dapat membaca file ".env": `, err)
	}
	/*
	TimeZone = os.Getenv("TIME_ZONE")
	var err error
	Loc, err = time.LoadLocation(TimeZone)
	if err != nil {
		log.Fatal(`seting location error: `, err)
	}
	*/
}