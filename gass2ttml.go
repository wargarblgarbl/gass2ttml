package main

import (
	"os"
	"bufio"
	"strings"
	"fmt"
	"strconv"
	"encoding/xml"
	"regexp"
	"flag"
	"io/ioutil"
)


/* Global varialbles and settings */

var inpath string
var outpath string

var language = flag.String("lang", "en", "default subtitle language")
var framerate = flag.String("frame", "24", "default framerate")


/* Various things to marshal our XML at the end */
var subtitles []*subtitle
var styles []*style
var regions []*region

type tt struct {
		Xmlns string `xml:"xmlns,attr"`
		XmlnsTtp string `xml:"xmlns:ttp,attr"`
		XmlnsTts string `xml:"xmlns:tts,attr"`
		XmlnsTtm string `xml:"xmlns:ttm,attr"`
		XmlnsXML string `xml:"xmlns:xml,attr"`
		TtpTimeBase string `xml:"ttp:timeBase,attr"`
		TtpFrameRate string `xml:"ttp:frameRate,attr"`
		XMLLang string `xml:"xml:lang,attr"`
		Head struct {
			Metadata struct {
				TtmTitle string `xml:"ttm:title"`
			} `xml:"metadata"`
			Styling struct {
				Style []style `xml:"style"`
			} `xml:"styling"`
			Layout struct {
				Region []region `xml:"region"`
			} `xml:"layout"`
		} `xml:"head"`
		Body struct {
			Region string `xml:"region,attr"`
			Style string `xml:"style,attr"`
			Div struct {
				P []subtitle `xml:"p"`
			} `xml:"div"`
		} `xml:"body"`
} 

type region struct {
	XMLID string `xml:"xml:id,attr"`
	TtsDisplayAlign string `xml:"tts:displayAlign,attr"`
	TtsExtent string `xml:"tts:extend,attr"`
	TtsOrigin string `xml:"tts:origin,attr"`
	
}

type style struct {
	XMLID string `xml:"xml:id,attr"`
	TtsTextAlign string `xml:"tts:textAlign,attr"`
	TtsFontFamily string `xml:"tts:fontFamily,attr"`
	TtsFontSize string `xml:"tts:fontSize,attr"`
}


type subtitle struct {
	Id string `xml:"xml:id,attr"`
	Begin string `xml:"begin,attr"`
	End string `xml:"end,attr"`
	Style string `xml:"style,attr,omitempty"`
	Region string `xml:"region,attr,omitempty"`
	Text string `xml:",innerxml"`
}

/* set our TTML options here */ 
func settt (v *tt) {
	v.Xmlns = "http://www.w3.org/ns/ttml"
	v.XmlnsTtp = "http://www.w3.org/ns/ttml#parameter"
	v.XmlnsTts = "http://www.w3.org/ns/ttml#styling"
	v.XmlnsTtm = "http://www.w3.org/ns/ttml#metadata"
	v.XmlnsXML = "http://www.w3.org/XML/1998/namespace"
	v.TtpTimeBase = "media"
}

func setopts (v *tt) {
	flag.Parse()
	/* set options and title */
	v.XMLLang = *language
	v.TtpFrameRate = *framerate
	v.Head.Metadata.TtmTitle = strings.Replace(inpath, ".ass", "", -1)

}

/* set remaining head options */
func sethead (v *tt){
	style := createstyle("s1", "center", "Arial", "100%")
	styles = append(styles, style)
	
	region1 := createregion("bottom", "after", "80% 40%", "10% 50%")
	regions = append(regions, region1)

	region2 := createregion("top", "before", "80% 40%", "10% 10%")
	regions = append(regions, region2)
	for _, a := range styles {
		v.Head.Styling.Style = append(v.Head.Styling.Style, *a)
	}
	for _, b := range regions {
		v.Head.Layout.Region = append(v.Head.Layout.Region, *b)
	}

}

/* set script defaults */
func setdefaults (v *tt) {
	v.Body.Region = "bottom"
	v.Body.Style = "s1"
}



func setsubtitles (v *tt) {
	loadass()
	for z, i := range subtitles {
		i.Id =  "subtitle"+strconv.Itoa(z+1)
		v.Body.Div.P = append(v.Body.Div.P, *i)
		
	}
}

/* Functions to create our objects */

func createline (begin string, end string, region string, text string) *subtitle {
	return &subtitle {
		Style : "s1",
		Begin : begin,
		End : end,
		Region : region,
		Text : text,
	}

}

func createregion (id string, align string, extent string, origin string) *region {
	return & region {
		XMLID : id,
		TtsDisplayAlign : align,
		TtsExtent : extent,
		TtsOrigin : origin,
	}
}

func createstyle (id string, align string, family string, size string) *style{
	return &style {
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

	
func procass (input string) {
	output := strings.Split(input, ",")
	var dialogue string
	if len(output) > 10 {
		dialogue = strings.Join(output[9:], ",")
		
	} else {
		dialogue = strings.Join(output[9:], "")
	}
	region := tagproc(dialogue)
//	dialogue = stripedit(dialogue)
	dialogue = striptags(dialogue)
	starttime := timeproc(output[1])
	endtime := timeproc(output[2])
	
	if dialogue != "" {
		z := createline(starttime, endtime, region, dialogue)
		subtitles = append(subtitles, z)
	}
}



func loadass() {
	fmt.Println(os.Args)
	fmt.Println(len(os.Args))	
	
	f, err := os.Open(inpath)
	if err != nil {
		fmt.Println("Cannot read file")
		os.Exit(1)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "Dialogue:") {
			procass(scanner.Text())
		}
	
	}
}




func main () {
	v := &tt{}
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

	
	out, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	bytehead := []byte(xml.Header)
	out = append(bytehead, out ...)
	f, arr := os.Create(outpath)
	if arr != nil {
		panic(arr)
	}
	defer f.Close()
	trr := ioutil.WriteFile(outpath, []byte(out), 0666)
	if trr != nil {
		panic(trr)
	}
}
	
	

