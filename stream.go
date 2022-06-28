package pkz

import (
	"archive/zip"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
)

const infoPath = "PKZ-INFO"
const zMe = 7575

type CStream struct {
	w io.Writer
}

func (c CStream) Write(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		p[i] ^= key
	}
	return c.w.Write(p)
}

func (c CStream) Close() error {
	return nil
}

const key = byte(170)

type DStream struct {
	r io.Reader
}

func (d DStream) Read(p []byte) (n int, err error) {
	n, err = d.r.Read(p)
	for i := 0; i < n; i++ {
		p[i] ^= key
	}
	return
}

func (d DStream) Close() error {
	return nil
}

type ComicInfoFetcher struct {
	ComicCount        func() (int, error)
	ComicInfo         func(comicIdx int) (*ComicInfo, error)
	ComicCover        func(comicIdx int, comicInfo *ComicInfo) ([]byte, error)
	ComicAuthorAvatar func(comicIdx int, comicInfo *ComicInfo) ([]byte, error)
	VolumeCount       func(comicIdx int, comicInfo *ComicInfo) (int, error)
	VolumeInfo        func(comicIdx int, comicInfo *ComicInfo, volumeIdx int) (*VolumeInfo, error)
	VolumeCover       func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo) ([]byte, error)
	ChapterCount      func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo) (int, error)
	ChapterInfo       func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int) (*ChapterInfo, error)
	ChapterCover      func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo) ([]byte, error)
	PictureCount      func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo) (int, error)
	PictureInfo       func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo, pictureIdx int) (*PictureInfo, error)
	PictureData       func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo, pictureIdx int, pictureInfo *PictureInfo) ([]byte, error)
}

func CreatePkz(
	srcW io.Writer,
	fetcher *ComicInfoFetcher,
) error {
	z := zip.NewWriter(srcW)
	z.RegisterCompressor(zMe, func(w io.Writer) (io.WriteCloser, error) {
		return CStream{
			w: w,
		}, nil
	})
	defer z.Close()
	comicCount, err := fetcher.ComicCount()
	if err != nil {
		return err
	}
	archive := Info{
		Comics:     make([]Comic, comicCount),
		ComicCount: comicCount,
	}

	for comicIdx := 0; comicIdx < len(archive.Comics); comicIdx++ {
		comicInfo, err := fetcher.ComicInfo(comicIdx)
		if err != nil {
			return err
		}
		archive.Comics[comicIdx] = Comic{
			ComicInfo: comicInfo,
		}
		// 漫画封面
		{
			cover, err := fetcher.ComicCover(comicIdx, comicInfo)
			if err != nil {
				return err
			}
			if cover != nil {
				path := uuid.New().String()
				w, err := z.Create(path)
				if err != nil {
					return err
				}
				_, err = w.Write(cover)
				if err != nil {
					return err
				}
				archive.Comics[comicIdx].CoverPath = path
			}
		}
		// 作者头像
		{
			aAva, err := fetcher.ComicAuthorAvatar(comicIdx, comicInfo)
			if err != nil {
				return err
			}
			if aAva != nil {
				path := uuid.New().String()
				w, err := z.Create(path)
				if err != nil {
					return err
				}
				_, err = w.Write(aAva)
				if err != nil {
					return err
				}
				archive.Comics[comicIdx].AuthorAvatarPath = path
			}
		}
		// 分卷
		volumesCount, err := fetcher.VolumeCount(comicIdx, comicInfo)
		if err != nil {
			return err
		}
		archive.Comics[comicIdx].Volumes = make([]Volume, volumesCount)
		archive.Comics[comicIdx].VolumesCount += volumesCount
		archive.VolumesCount += volumesCount
		for volumeIdx := 0; volumeIdx < len(archive.Comics[comicIdx].Volumes); volumeIdx++ {
			volumesInfo, err := fetcher.VolumeInfo(comicIdx, comicInfo, volumeIdx)
			if err != nil {
				return err
			}
			archive.Comics[comicIdx].Volumes[volumeIdx] = Volume{
				VolumeInfo: volumesInfo,
			}
			// 分卷封面
			{
				cover, err := fetcher.VolumeCover(comicIdx, comicInfo, volumeIdx, volumesInfo)
				if err != nil {
					return err
				}
				if cover != nil {
					path := uuid.New().String()
					w, err := z.Create(path)
					if err != nil {
						return err
					}
					_, err = w.Write(cover)
					if err != nil {
						return err
					}
					archive.Comics[comicIdx].Volumes[volumeIdx].CoverPath = path
				}
			}
			// 章节
			chapterCount, err := fetcher.ChapterCount(comicIdx, comicInfo, volumeIdx, volumesInfo)
			if err != nil {
				return err
			}
			archive.Comics[comicIdx].Volumes[volumeIdx].Chapters = make([]Chapter, chapterCount)
			archive.Comics[comicIdx].Volumes[volumeIdx].ChapterCount += chapterCount
			archive.Comics[comicIdx].ChapterCount += chapterCount
			archive.ChapterCount += chapterCount
			for chapterIdx := 0; chapterIdx < len(archive.Comics[comicIdx].Volumes[volumeIdx].Chapters); chapterIdx++ {
				chapterInfo, err := fetcher.ChapterInfo(comicIdx, comicInfo, volumeIdx, volumesInfo, chapterIdx)
				if err != nil {
					return err
				}
				archive.Comics[comicIdx].Volumes[volumeIdx].Chapters[chapterIdx] = Chapter{
					ChapterInfo: chapterInfo,
				}
				// 图片
				pictureCount, err := fetcher.PictureCount(comicIdx, comicInfo, volumeIdx, volumesInfo, chapterIdx, chapterInfo)
				if err != nil {
					return err
				}
				archive.Comics[comicIdx].Volumes[volumeIdx].Chapters[chapterIdx].Pictures = make([]Picture, pictureCount)
				archive.Comics[comicIdx].Volumes[volumeIdx].Chapters[chapterIdx].PictureCount += pictureCount
				archive.Comics[comicIdx].Volumes[volumeIdx].PictureCount += pictureCount
				archive.Comics[comicIdx].PictureCount += pictureCount
				archive.PictureCount += pictureCount
				for pictureIdx := 0; pictureIdx < len(archive.Comics[comicIdx].Volumes[volumeIdx].Chapters[chapterIdx].Pictures); pictureIdx++ {
					pictureInfo, err := fetcher.PictureInfo(comicIdx, comicInfo, volumeIdx, volumesInfo, chapterIdx, chapterInfo, pictureIdx)
					if err != nil {
						return err
					}
					archive.Comics[comicIdx].Volumes[volumeIdx].Chapters[chapterIdx].Pictures[pictureIdx] = Picture{
						PictureInfo: pictureInfo,
					}
					{
						cover, err := fetcher.PictureData(comicIdx, comicInfo, volumeIdx, volumesInfo, chapterIdx, chapterInfo, pictureIdx, pictureInfo)
						if err != nil {
							return err
						}
						if cover != nil {
							path := uuid.New().String()
							w, err := z.Create(path)
							if err != nil {
								return err
							}
							_, err = w.Write(cover)
							if err != nil {
								return err
							}
							archive.Comics[comicIdx].Volumes[volumeIdx].Chapters[chapterIdx].Pictures[pictureIdx].PicturePath = path
						}
					}
				}
			}
		}
	}
	buff, err := json.Marshal(&archive)
	if err != nil {
		w, err := z.Create(infoPath)
		if err != nil {
			return err
		}
		_, err = w.Write(buff)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadPkz(filePath string, innerPath string) ([]byte, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	r.RegisterDecompressor(zMe, func(r io.Reader) io.ReadCloser {
		return DStream{
			r: r,
		}
	})
	re, err := r.Open(innerPath)
	if err != nil {
		return nil, err
	}
	defer re.Close()
	return ioutil.ReadAll(re)
}

func ReadPkzInfo(filePath string) (*Info, error) {
	buff, err := ReadPkz(filePath, infoPath)
	if err != nil {
		return nil, err
	}
	var info Info
	err = json.Unmarshal(buff, &info)
	if err != nil {
		return nil, err
	}
	return &info, err
}
