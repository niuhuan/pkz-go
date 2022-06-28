package pkz

import (
	"archive/zip"
	"compress/flate"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
)

const infoPath = "PKZ-INFO"
const zMe = zip.Deflate
const key = byte(170)

type CStream struct {
	w *flate.Writer
}

func (c *CStream) Write(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		p[i] ^= key
	}
	return c.w.Write(p)
}

func (c *CStream) Close() error {
	return c.w.Close()
}

func NewCStream(wr io.Writer) (io.WriteCloser, error) {
	w, e := flate.NewWriter(wr, flate.BestCompression)
	if e != nil {
		return nil, e
	}
	return &CStream{
		w: w,
	}, nil
}

type DStream struct {
	r io.ReadCloser
}

func (d *DStream) Read(p []byte) (n int, err error) {
	n, err = d.r.Read(p)
	for i := 0; i < n; i++ {
		p[i] ^= key
	}
	return
}

func (d *DStream) Close() error {
	return d.r.Close()
}

func NewDStream(r io.Reader) io.ReadCloser {
	return &DStream{
		r: flate.NewReader(r),
	}
}

type ComicsFetcher struct {
	ArchiveInfo         func() (*ArchiveInfo, error)
	ArchiveCover        func() ([]byte, error)
	ArchiveAuthorAvatar func() ([]byte, error)
	ComicCount          func() (int, error)
	ComicInfo           func(comicIdx int) (*ComicInfo, error)
	ComicCover          func(comicIdx int, comicInfo *ComicInfo) ([]byte, error)
	ComicAuthorAvatar   func(comicIdx int, comicInfo *ComicInfo) ([]byte, error)
	VolumeCount         func(comicIdx int, comicInfo *ComicInfo) (int, error)
	VolumeInfo          func(comicIdx int, comicInfo *ComicInfo, volumeIdx int) (*VolumeInfo, error)
	VolumeCover         func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo) ([]byte, error)
	ChapterCount        func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo) (int, error)
	ChapterInfo         func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int) (*ChapterInfo, error)
	ChapterCover        func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo) ([]byte, error)
	PictureCount         func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo) (int, error)
	PictureInfo          func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo, pictureIdx int) (*PictureInfo, error)
	PictureData          func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo, pictureIdx int, pictureInfo *PictureInfo) ([]byte, error)
}

func WritePkz(
	srcW io.Writer,
	fetcher *ComicsFetcher,
) error {
	z := zip.NewWriter(srcW)
	z.RegisterCompressor(zMe, func(w io.Writer) (io.WriteCloser, error) {
		return NewCStream(w)
	})
	defer z.Close()
	archiveInfo, err := fetcher.ArchiveInfo()
	if err != nil {
		return err
	}
	comicCount, err := fetcher.ComicCount()
	if err != nil {
		return err
	}
	archive := Archive{
		ArchiveInfo: archiveInfo,
		Comics:      make([]Comic, comicCount),
		ComicCount:  comicCount,
	}
	{
		aAva, err := fetcher.ArchiveCover()
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
			archive.CoverPath = path
		}
	}
	{
		aAva, err := fetcher.ArchiveAuthorAvatar()
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
			archive.AuthorAvatarPath = path
		}
	}
	// comics
	for comicIdx := 0; comicIdx < len(archive.Comics); comicIdx++ {
		comicInfo, err := fetcher.ComicInfo(comicIdx)
		if err != nil {
			return err
		}
		archive.Comics[comicIdx] = Comic{
			ComicInfo: comicInfo,
			Idx:       comicIdx,
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
				Idx:        volumeIdx,
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
					Idx:         chapterIdx,
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
						Idx:         pictureIdx,
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
		return err
	}
	w, err := z.Create(infoPath)
	if err != nil {
		return err
	}
	_, err = w.Write(buff)
	return err
}

func ReadPkzPath(filePath string, innerPath string) ([]byte, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	r.RegisterDecompressor(zMe, func(r io.Reader) io.ReadCloser {
		return NewDStream(r)
	})
	re, err := r.Open(innerPath)
	if err != nil {
		return nil, err
	}
	defer re.Close()
	return ioutil.ReadAll(re)
}

func ReadPkzArchive(filePath string) (*Archive, error) {
	buff, err := ReadPkzPath(filePath, infoPath)
	if err != nil {
		return nil, err
	}
	var info Archive
	err = json.Unmarshal(buff, &info)
	if err != nil {
		return nil, err
	}
	return &info, err
}
