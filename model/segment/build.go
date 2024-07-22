package segment

func BuildTextSegment(text string) *Segment {
	return &Segment{
		Type: "text",
		Data: &Segment_Data{
			Text: text,
		},
	}
}

func BuildFaceSegment(id string) *Segment {
	return &Segment{
		Type: "face",
		Data: &Segment_Data{
			Id: id,
		},
	}
}

func BuildImageSegment(file string) *Segment {
	return &Segment{
		Type: "image",
		Data: &Segment_Data{
			File: file,
		},
	}
}

func BuildAtSegment(qq string) *Segment {
	return &Segment{
		Type: "at",
		Data: &Segment_Data{
			Qq: qq,
		},
	}
}

func BuildShareSegment(url, title string) *Segment {
	return &Segment{
		Type: "share",
		Data: &Segment_Data{
			Url:   url,
			Title: title,
		},
	}
}
