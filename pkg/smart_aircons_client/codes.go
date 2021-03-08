package smart_aircons_client

import (
	"fmt"
	"log"
)

var (
	fujitsuCodes = map[string]string{
		"off":      "1:1,0,37000,1,1,121,60,16,14B16,45BCBBBCCBBBCCBBBBBB16,17,13,16BBBBBCBBBBBBBCBBBBCBBBBBB16,0",
		"cool_18":  "1:1,0,37000,1,1,121,61,15,16B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBCBBBBCBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCBCBC14,0",
		"cool_19":  "1:1,0,37000,1,1,122,61,15,16B15,45BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCCBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCCBBC16,0",
		"cool_20":  "1:1,0,37000,1,1,122,60,16,15B15,45BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBBCBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCBBBC16,0",
		"cool_21":  "1:1,0,37000,1,1,122,61,15,15B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCBCBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCCCCB15,0",
		"cool_22":  "1:1,0,37000,1,1,122,60,15,16B15,45BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBCCBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCBCCB15,0",
		"cool_23":  "1:1,0,37000,1,1,122,61,15,16B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCCCBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCCBCB16,0",
		"cool_24":  "1:1,0,37000,1,1,123,62,15,16B16,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBBBCCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCBBCB15,0",
		"cool_25":  "1:1,0,37000,1,1,123,60,16,16B15,46BCBBBCCBBBCCBBBBB13,16BBBBBBBCBBBBBBBCBBBBDCCCCCCDBBCBBBBBBBBCCBBBBBBCBBCCBBBBBBBBBBBBBBBBBBDDBBBBBBBBBBBBBBBBBBBCCCCCCBB15,0",
		"cool_26":  "1:1,0,37000,1,1,122,61,15,17B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBCBCCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCBCBB15,0",
		"cool_27":  "1:1,0,37000,1,1,122,61,15,16B14,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCCBCCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCCBBB15,0",
		"cool_28":  "1:1,0,37000,1,1,123,62,15,15B15,45BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBBCCCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCBBBB15,0",
		"cool_29":  "1:1,0,37000,1,1,123,62,15,15B14,47BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCBCCCBBBBBBBBBBBBBBBBBBBBBBBBBB15,19,12,16BBBBBBBBBBBCCCCCCCC15,0",
		"cool_30":  "1:1,0,37000,1,1,122,61,15,16B15,48BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBCCCCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCBCCC15,0",
		"heat_16":  "1:1,0,37000,1,1,122,61,15,15B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBCBBBBBBBBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCBCBBCC15,0",
		"heat_17":  "1:1,0,37000,1,1,121,62,15,16B16,45BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCBBBBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCBC16,0",
		"heat_18":  "1:1,0,37000,1,1,122,61,15,15B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBCBBBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCBCBC14,0",
		"heat_19":  "1:1,0,37000,1,1,123,60,15,16B14,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCCBBBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCBBC15,0",
		"heat_20":  "1:1,0,37000,1,1,121,62,15,16B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBBCBBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCBBBC15,0",
		"heat_21":  "1:1,0,37000,1,1,122,61,15,16B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCBCBBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCCB15,0",
		"heat_22":  "1:1,0,37000,1,1,123,61,15,16B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBCCBBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCBCCB15,0",
		"heat_23":  "1:1,0,37000,1,1,123,61,15,15B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCCCBBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCBCB15,0",
		"heat_24":  "1:1,0,37000,1,1,124,60,16,14B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBBBCBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCBBCB15,0",
		"heat_25":  "1:1,0,37000,1,1,122,61,15,16B14,45BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCBBCBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCBB16,0",
		"heat_26":  "1:1,0,37000,1,1,122,61,15,15B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBCBCBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCBCBB16,0",
		"heat_27":  "1:1,0,37000,1,1,123,60,16,14B16,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCCBCBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCBBB15,0",
		"heat_28":  "1:1,0,37000,1,1,123,60,16,16B15,45BCBBB14,48CBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCDCCCCBBBCBBBBBBBBDCBBBBBBBBCCBBCBB13,15BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCBBBB15,0",
		"heat_29":  "1:1,0,37000,1,1,123,59,16,14B15,46BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBCBCCBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCCCCC16,0",
		"heat_30":  "1:1,0,37000,1,1,123,60,15,14B15,45BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBBBBBBCCCBBCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCBCCC15,0",
		"fan_only": "1:1,0,37000,1,1,123,59,16,15B16,45BCBBBCCBBBCCBBBBBBBBBBBBBCBBBBBBBCBBBBBCCCCCCBBBCBBBBBBBBCCBBCBBBBCCCCCBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBCCBCCC16,0",
	}

	mitsubishiCodes = map[string]string{
		"off":      "1:1,0,37000,1,1,126,63,16,47B16,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCCCCCBCCCCCBCCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBCBCBB16,631,131,65BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCCCCCBCCCCCBCCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBCBCBB15,0",
		"cool_16":  "1:1,0,37000,1,1,126,64,15,49B16,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCCCCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCBBBBB15,629,132,63BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCCCCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCBBBBB15,0",
		"cool_17":  "1:1,0,37000,1,1,126,65,15,49B16,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCC13,17BCCCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBBBB16,633,132,65BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBCCCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBBBB15,0",
		"cool_18":  "1:1,0,37000,1,1,128,64,15,49B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCBCCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCBBBBBB15,637,131,63BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCBCCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCBBBBBB15,0",
		"cool_19":  "1:1,0,37000,1,1,128,63,16,49B15,16CCBCCBBCBCCBBCBBCCBCCBCCC18,15CCCCCCCCCCCCCCCCBCCCCCBBCCCBBCCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBBBBB16,631,131,63BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBBCCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBBBBB15,0",
		"cool_20":  "1:1,0,37000,1,1,128,64,15,47B15,16CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCCBCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBBBBBB16,633,131,64BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCCBCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBBBBBB15,0",
		"cool_21":  "1:1,0,37000,1,1,126,65,15,48B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBCBCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC15,629,134,64BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBCBCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC15,0",
		"cool_22":  "1:1,0,37000,1,1,126,65,15,49B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCBBCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCCCCCC15,629,132,65BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCBBCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCCCCCC15,0",
		"cool_23":  "1:1,0,37000,1,1,126,63,16,47B15,16CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBBBCCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCCCCC15,630,132,65BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBBBCCCCCCBBCBBCCCCC16,50BBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCCCCC15,0",
		"cool_24":  "1:1,0,37000,1,1,126,64,15,48B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCCCBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCCCCC15,631,131,65BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCCCBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCCCCC15,0",
		"cool_25":  "1:1,0,37000,1,1,126,65,15,48B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBCCBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCC19,15CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCCCC15,631,131,64BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBCCBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCCCC15,0",
		"cool_26":  "1:1,0,37000,1,1,127,65,15,49B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCBCBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCBCCCCC15,631,131,65BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCBCBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCBCCCCC15,0",
		"cool_27":  "1:1,0,37000,1,1,127,63,16,49B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBBCBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCCCC15,631,131,64BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBBCBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCCCC16,0",
		"cool_28":  "1:1,0,37000,1,1,126,65,15,49B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCCBBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBCCCCC15,631,131,64BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCCBBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBCCCCC15,0",
		"cool_29":  "1:1,0,37000,1,1,127,65,15,49B16,16CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBCBBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCCC15,631,131,65BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBCBBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCCC15,0",
		"cool_30":  "1:1,0,37000,1,1,127,64,15,48B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCBBBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCBCCCC15,629,132,65BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCBBBCCCCCBBCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCBCCCC15,0",
		"cool_31":  "1:1,0,37000,1,1,126,65,15,48B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBBBBCCCCCBBCBBCCCCCBBBCBCCCCCCCCCCCCC18,15CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCBCCBC15,631,131,65BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBBBBCCCCCBBCBBCCCCCBBBCBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCBCCBC15,0",
		"heat_16":  "1:1,0,37000,1,1,126,64,16,49B15,16CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBCCCCCCCCCCCCCCCCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCBCCBBB15,631,131,65BBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBCCCCCCCCCCCCCCCCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCBCCBBB15,0",
		"heat_17":  "1:1,0,37000,1,1,124,64,15,49B14,20,13,17DBCD15,52,12,48CBCCBBC12,52FDDBCCBCDDDDDCDDCCDCDDDDC17,15DBCCCCDBCDDDBCDCDDDDDCCDBBCCDCCBBBBCCDHDCDDDCDCCDDDDCDDDDDDDDDCDCCCDDDCDDDCDCCDDDCDCDCCCDHDDCBBDHBBB14,634,130,67FBCCDBDDEFDBDDBBCBBDCBCDBDDCCDCDCDCCDDDCDCDCDBDCDCDBCDDDBDHDHDDCDCDDBBDDCDDEFBBDCDDCDHHDDCCDDDDDCCCDDDDCCHCCDCDCCCDCDDDDDDDDCDDDCDCDCCDDCBBCDBBB14,0",
		"heat_18":  "1:1,0,37000,1,1,127,64,15,49B15,17C13,20,12,49CCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCD12,17CCC13,52FCCCCBCCCCCBCCCCCCCCC16,20EBCDFCCBGBBCCCCCCCCCCCDFCCCCCCCCCCCCCCCCCDFCDDCDDDCCDDDFCCCCCCCCCCCCBBBCCBBB15,633,130,66BBCCCBCCBBCBCCBBDEBCCBCCBCCCDFCCCCCDFCCCCCDFCBDDDCCBCCCCCBCCCCCCCCCCBBCDDCCBBBGDCCCCCCCDFCCDFCCCCDFDFCCCCCDDFCCCCCCCDDFCCCCCDFCCCCCCCCCCBBBCCBBB15,0",
		"heat_19":  "1:1,0,37000,1,1,126,65,15,49B14,18CCBCCBBC12,49CCBBCDBCCDCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBCCCCB13,52CCCCCCCCCCDBCCCCCEDBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCDCBBB13,640ABECCCBCCBBCBCCBBCBBCCDCCBCCCC14,21,11,18CCCCCC17,20CCCCCCCBCCCCCBCCCCDBCCCCCCCCCCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCEHBBB13,0",
		"heat_20":  "1:1,0,37000,1,1,126,65,15,49B15,17CCBC13,20,12,49BCBDDEBDEBCCBCCBCCCCCCCCDDCCCCCCD12,17CCBCCCCC13,52FCCCCCBCCCCCCCCDEGFCCCCBBGEDFCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCDDCCCCDFCCCCCCCCCDDDCCCGFCBCBBB15,642,120,65BBCDCECCBGFBCCBBCBBCCBDFBCCCCCCCDFCCCCCCCCCCCBCDFCDGFCCCC16,22,8,49CCCCCCCDFBBCCCCCBBBBCDFCCCCCCCCDFCCCCCCDFCCCCCDFCCCCCCCCDDCCCDFCCCCCCCCDDFCCCBDDBCBBB13,0",
		"heat_21":  "1:1,0,37000,1,1,126,65,15,49B15,17CCBC15,20,12,49BCECCBBCBBCCBCCBCCCD12,17DFCCDFCCCCCCCCDBCCCCC15,52FCCCBCBCCCCCCCCCBGFCCCD12,52EBBCCCCCDFCCCCCCCDFCCCCCCCCCCCCDFCCCCCCCDFDFCCCDFCCCCCD12,20FCDFCBCBCBBG12,633,129,65BBCCCBCCBBCBCCGECBBDFBCCBCCCCCCCDFDFCCDFCCCDFBCCCCCBCCCCBCBCCCCCCCCCBBCCCCCBBBBCCIFCCCDFCDFCCCCCCCCCCCCCCCCCCDFCCCCCCCCDIFCCCCCCCD8,18DFCCCCBCBCGEB15,0",
		"heat_22":  "1:1,0,37000,1,1,126,65,15,49B15,17CCBCCBBC15,52,12,17CBDEBBCCBCCBCCCCCCCCE15,20ECCFECCCCF12,50CCCCCDEFECCDGCCCCCCCCFGBCCCFEDGBDECCCCCCCCCFECCCCCCCCFECCCCCCCCCCCCCFECCCCCCCCCCCCCCCCCCCCBBCBCBBB15,633,129,68GBFECBCCBDEBCCBBFBBCCBCCBCCCCCFECCCCCFECCCCCCBCCCCCBCCCCCBBFECCCCFECBBCCCCFGBBBCCCCFECCFEFECCCF12,20ECCCCCCCCFECCCCCCCFECCCCCCCCCCCCCCCCCCCCBBCBFGGB13,0",
		"heat_23":  "1:1,0,37000,1,1,127,64,15,51,13,48,16,17,14,21,12,17BDDBBDBEECBDBBEEBDDBFFDDFDDDDEFDFDDEEEFDBDFDFDBDFDDBBBDEEEFDDDECBDDDDDBBBBEDDDDEFDDDDDDDDDDDDEFDDDDDDDDDDDFDDDEFDDFEFDDDDDDDFDDDDFDEFB16,48DBBB13,640,121,65BBEFDCDDBBDBDEBBDBBDEBDDBFDEFDDDDDFDFDDDDDFDDBDDDDDBDDDFBBBDDDDEFDDDBBFDEEEBBBBDDDDDDFDDDDFEFDDDDDDDEEFDDDDDFEFDDDDEFDDDDDEDDDDDDDEFDDDDDDBCDBBC13,0",
		"heat_24":  "1:1,0,37000,1,1,126,65,15,52,12,49,14,21,12,17,15,17,15,49FFBCFGFDCGFGGFFBFDCFDEFFFFEFFEFFDEEDEFDCFFFFFGFDEFEDDBDDEFDEDDCBEFFDDCGBCFDEFDDEDEFFFFDEFFFDEDDEEFDEFFFDEFFDEDEFFDEDEFFDEEEFFEDEFFBEGBEBBG15,633AGGFFEGEDBCDCDEBCDCGDEGFFGDEFDEEDEFFFEFDDEFFDEGEDEDEGFFFFFEFGFFFFFEDDBCDEFDEBBCGFEDEEEFFFFFFFEFFDEDEFFEFFFEEFFFDEDEFFFEFEFDEDEFDFDEFFFFFFBEBCFGGB15,0",
		"heat_25":  "1:1,0,37000,1,1,125,64,16,48B15,18CCBCCBBCBCCBBCBBCCBC14,21,12,49CCCCCCCCCCCCCCCCCC12,18C15,52FCCCCBCCFCBCCBCCCCCCCCEBCCCCCBBBBCCCCCCCCCDCCFCCCCFCCCCCCCCCFCCCCCCCCCCCCCCCC17,15CCCCCCCCFCCCCBBBFEEB15,633,130,64BECCCECCBECBCDEBCEECCEDFBCCCCCCCDFCCCCCCCCCCCECCCCFBCCCFBCCBCHCCCCCCBBCFCCCBBBBCCCFCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBCBBB15,0",
		"heat_26":  "1:1,0,37000,1,1,127,64,15,49B13,19CCBCCBBCBCC15,52,12,49CBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCEC16,15CCCBCBCCCCCFCCBDCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC16,19CCCCCCCCCCCCCCCCCCGCCCBBDECEBB15,638,130,64DECCCBFCBDCDCCBDCBBCCECCBCCCCCCCCCCCCCCCCCCFCBCCCCCECCCCCBCBCCCCCCCCDECCCCCBDEBCCCCCCCCCCCCCCCCCCCCCCCCCCGCCCCCCCCCCCCCCCCCCCCCFCCCCCCCCBDEBCBBB15,0",
		"heat_27":  "1:1,0,37000,1,1,126,65,15,49B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCC11,18CCCCCC14,20CCBEDCCC15,52DCCE12,49BCBCCCCCCCCBBCCCCCBBBBCECCCCCCCCCCCCCCCCCCCCCCECEDCCCCCCCCCCCCCECDCCCCCCCCCCCCCCCCCBBBB15,633,129,66BBCCCBCCBBCBCEGBCBBCCBCCBCCEDDCCCECDCCCCCCCCCBCDDCDBCCCCBBCBCCCCCCCE12,52GCCECCBBBBCCEDCCCCCCCCCEDCCCCCCCCECCCCCCCEDCCCCCCCCCCCCCCCCCCEDCCCDDCCCBBBB13,0",
		"heat_28":  "1:1,0,37000,1,1,126,65,15,49B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBCCCCCCBBCCCCCCCCBBCCCCCBBBBCCCCCCCCC15,20,12,17CCCCCCCCCCCCCCCCCCCCCCCCCCCDECCCCCCCCCCCCDECCCBCCCBBBB14,634,129,65BBCCD12,49CCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCDECCCBD12,20ECCBCCCCCCBBCDECCECCBBDECCCBBB15,52ECDCCCCCCCCCCECCCDECCCCCCCCCCCCDEDECCCCCCCCCCDECDECCCCCCDHDCCBJHB15,0",
		"heat_29":  "1:1,0,37000,1,1,127,64,15,49B15,17CCBCCBBCBCC12,51BCBBCC15,52,11,20CBCCCCCCCCCCCCCCCCFFCCBCCCCCDC14,20CCBCBBCCCCCCCCBBCCCG12,17BBBBGFCCCCCCCCCCCCCFCCCFFHCCCCCCCCCGHCFCCCCCCCCCGHGHCGCCCCCCGFBCCBBBB15,633AEDCCGBCCBBGDCGBBCBBCCBCCBCCGHCCCCCGFCCCCCCCCCBCCCCCBGHCCBCBBCCCCCCCCBBCCCCFDBBBCCGFCCCCCGFCCCFCCCCCCGFHCGHCCCCCCCCCCCCCCCCCCCCCCCFFHCGCCCBCCBBBB14,0",
		"heat_30":  "1:1,0,37000,1,1,127,65,15,49B16,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBCCCCCBBBCCCCCCCCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCBBBB16,636ABBCCCBCCBBCBCCBBCBBCCBCCBCCC19,15CCCCCCCCCCCCCCCC15,52,12,17CCCCBCCCCCBBBCCCCCCCCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCBBBB15,0",
		"heat_31":  "1:1,0,37000,1,1,127,65,15,49B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBCCCCBBBBCCCCCCCCBBCCCCCBBBCBCCCCCCCCCCCCCCCCCCC14,20CCCCCCCCCCCCCCCCCCCCCCCCCD12,17CCCCCCCCCCCBD12,49BCC15,636ABBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBCCCCBBBBCCCCCCCCBBCCCCCBBBCBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCDECCCCCCCCCCCCCCCCCCCCCCCCBCBBCC14,0",
		"fan_only": "1:1,0,37000,1,1,126,65,15,49B15,17CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCCBCCCCCCBCCCCCBCCBBCCCCCBBBBCCCCCCCCCCCCC15,22,10,17CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBCBBBB15,637ABBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCCBCCCCCCBCCCCCBCCBBCCCCCBBBBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBBCBBBB15,0",
	}

	allCodes = map[string]map[string]string{
		"fujitsu":    fujitsuCodes,
		"mitsubishi": mitsubishiCodes,
	}
)

func GetCode(name string, on bool, mode string, temperature int64) (string, error) {
	var ok bool
	var codes map[string]string
	var code string

	codes, ok = allCodes[name]
	if !ok {
		return "", fmt.Errorf("%#+v not a recognized name", name)
	}

	var codeName = ""

	if !on || mode == "off" {
		codeName = "off"
	} else if mode == "fan_only" {
		codeName = "fan_only"
	} else {
		codeName = fmt.Sprintf("%v_%v", mode, temperature)
	}

	code, ok = codes[codeName]
	if !ok {
		return "", fmt.Errorf("%#+v not a recognized code for %#+v", code, name)
	}

	log.Printf("name=%#+v, codeName=%#+v, on=%#+v, mode=%#+v, temperature=%#+v, code=%#+v", name, codeName, on, mode, temperature, code)

	return code, nil
}
