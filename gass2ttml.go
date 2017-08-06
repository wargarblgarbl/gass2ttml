package main

import (
	"os"
	"strings"
	"fmt"
	"strconv"
//	"encoding/xml"
	"regexp"
	"flag"
//	"io/ioutil"
	"github.com/wargarblgarbl/libgosubs/ass"
	"github.com/wargarblgarbl/libgosubs/ttml"
)


/* Global varialbles and settings */

var inpath string
var outpath string

var language = flag.String("lang", "en", "default subtitle language")
var framerate = flag.String("frame", "24", "default framerate")


/* Various things to marshal our XML at the end */
var subtitles []*ttml.Subtitle

/* set our TTML options here */ 
func settt (v *ttml.WTt) {
	v.Xmlns = "http://www.w3.org/ns/ttml"
	v.XmlnsTtp = "http://www.w3.org/ns/ttml#parameter"
	v.XmlnsTts = "http://www.w3.org/ns/ttml#styling"
	v.XmlnsTtm = "http://www.w3.org/ns/ttml#metadata"
	v.XmlnsXML = "http://www.w3.org/XML/1998/namespace"
	v.TtpTimeBase = "media"
}

func setopts (v *ttml.WTt) {
	flag.Parse()
	/* set options and title */
	v.XMLLang = *language
	v.TtpFrameRate = *framerate
	v.Head.Metadata.TtmTitle = strings.Replace(inpath, ".ass", "", -1)

}

/* set remaining head options */
func sethead (v *ttml.WTt){
	style := createstyle("s1", "center", "Arial", "100%")
	v.Head.Styling.Style = append(v.Head.Styling.Style, *style)
	region1 := createregion("bottom", "after", "80% 40%", "10% 50%")
	region2 := createregion("top", "before", "80% 40%", "10% 10%")
	v.Head.Styling.Style = append(v.Head.Styling.Style, *style)
	v.Head.Layout.Region = append(v.Head.Layout.Region, *region1)
	v.Head.Layout.Region = append(v.Head.Layout.Region, *region2)

}

/* set script defaults */
func setdefaults (v *ttml.WTt) {
	v.Body.Region = "bottom"
	v.Body.Style = "s1"
}



func setsubtitles (v *ttml.WTt) {
	loadass()
	//Only reason to do this atm is because ass does not have a subtitle ID.
	//Potentially change this in the future. 
	for z, i := range subtitles {
		i.Id =  "subtitle"+strconv.Itoa(z+1)
		v.Body.Div.P = append(v.Body.Div.P, *i)
		
	}
}

/* Functions to create our objects */

func createline (begin string, end string, region string, text string) *ttml.Subtitle {
	return &ttml.Subtitle {
		Style : "s1",
		Begin : begin,
		End : end,
		Region : region,
		Text : text,
	}

}

func createregion (id string, align string, extent string, origin string) *ttml.Region {
	return &ttml.Region {
		XMLID : id,
		TtsDisplayAlign : align,
		TtsExtent : extent,
		TtsOrigin : origin,
	}
}

func createstyle (id string, align string, family string, size string) *ttml.Style{
	return &ttml.Style {
		XMLID : id,
		TtsTextAlign : align,
		TtsFontFamily : family,
		TtsFontSize : size,
	}
}




/*working with ass and editor tags */
func striptags (input string)(output string) {
	if strings.ContainsAny(input, "{ & }") == false {
		output = input
	} else {
		r := regexp.MustCompile(`{[^}]*}`)
		matches := r.FindAllString(input, -1)
		proc := input
		for _, i := range matches {
			proc = strings.Replace(proc, i, "", -1)
		}
		output = proc
	}
	return
}

func stripedit (input string)(output string) {
	string := strings.Split(input, "//")
	output = string[0]
	return
}

func tagproc (input string)(region string) {
	if strings.Contains(input, `\an8`) {
		region = "top"
	}
	return

}

/* Pad the timestamps, if we're in the multi-hour range, pad only the right hand side. */

func timeproc (input string)(output string) {
	if len(input) == 10 {
		input = "0"+input+"0"
	} else if len(input) >= 11 {
		input = input+"0"
	}
	output = input
	return
	
}


func loadass() {
	fmt.Println(os.Args)
	fmt.Println(len(os.Args))	
	loadedass := ass.ParseAss(inpath)
	for _, i := range loadedass.Events.Body {
		if i.Format == "Dialogue" {
			region := tagproc(i.Text)
			//dialogue = stripedit(dialogue)
			dialogue := striptags(i.Text)
			starttime := timeproc(i.Start)
			endtime := timeproc(i.End)
			if dialogue != "" {
				z := createline(starttime, endtime, region, dialogue)
				subtitles = append(subtitles, z)
			}
		}
	}

}




func main () {
	v := &ttml.WTt{}
	if len(os.Args) > 1  {
		inpath = os.Args[1]
	} else {
		fmt.Println("Usage : " + os.Args[0] + " input_file_name output_file_name")
		os.Exit(1)
	}
	if len(os.Args) > 2 {
		outpath = os.Args[2]
	} else {
		outpath = strings.Replace(inpath, "ass", "xml", -1)
	}
	setsubtitles(v)
	setdefaults(v)
	settt(v)
	setopts(v)
	sethead(v)
	ttml.WriteTtml(v, outpath)
}
	

