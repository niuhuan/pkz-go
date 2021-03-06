package pkz

type Archive struct {
	*ArchiveInfo
	CoverPath        string  `json:"cover_path"`
	AuthorAvatarPath string  `json:"author_avatar_path"`
	Comics           []Comic `json:"comics"`
	ComicCount       int     `json:"comic_count"`
	VolumesCount     int     `json:"volumes_count"`
	ChapterCount     int     `json:"chapter_count"`
	PictureCount     int     `json:"picture_count"`
}

type ArchiveInfo struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

type Comic struct {
	*ComicInfo
	CoverPath        string   `json:"cover_path"`
	AuthorAvatarPath string   `json:"author_avatar_path"`
	Volumes          []Volume `json:"volumes"`
	VolumesCount     int      `json:"volumes_count"`
	ChapterCount     int      `json:"chapter_count"`
	PictureCount     int      `json:"picture_count"`
	Idx              int      `json:"idx"`
}

type Volume struct {
	*VolumeInfo
	CoverPath    string    `json:"cover_path"`
	Chapters     []Chapter `json:"chapters"`
	ChapterCount int       `json:"chapter_count"`
	PictureCount int       `json:"picture_count"`
	Idx          int       `json:"idx"`
}

type Chapter struct {
	*ChapterInfo
	CoverPath    string    `json:"cover_path"`
	Pictures     []Picture `json:"pictures"`
	PictureCount int       `json:"picture_count"`
	Idx          int       `json:"idx"`
}

type Picture struct {
	*PictureInfo
	PicturePath string `json:"picture_path"`
	Idx         int    `json:"idx"`
}

type ComicInfo struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Categories  []string `json:"categories"`
	Tags        []string `json:"tags"`
	AuthorId    string   `json:"author_id"`
	Author      string   `json:"author"`
	UpdatedAt   int64    `json:"updated_at"`
	CreatedAt   int64    `json:"created_at"`
	Description string   `json:"description"`
	ChineseTeam string   `json:"chinese_team"`
	Finished    bool     `json:"finished"`
}

type VolumeInfo struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	UpdatedAt int64  `json:"updated_at"`
	CreatedAt int64  `json:"created_at"`
}

type ChapterInfo struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	UpdatedAt int64  `json:"updated_at"`
	CreatedAt int64  `json:"created_at"`
}

type PictureInfo struct {
	Id     string `json:"id"`
	Title  string `json:"title"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format"`
}
