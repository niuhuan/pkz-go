package pkz

import (
	"archive/zip"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
)

const infoPath = "PKZ-INFO"
const key = byte(170)

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
	PictureCount        func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo) (int, error)
	PictureInfo         func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo, pictureIdx int) (*PictureInfo, error)
	PictureData         func(comicIdx int, comicInfo *ComicInfo, volumeIdx int, volumeInfo *VolumeInfo, chapterIdx int, chapterInfo *ChapterInfo, pictureIdx int, pictureInfo *PictureInfo) ([]byte, error)
}

func WritePkz(
	srcW io.Writer,
	fetcher *ComicsFetcher,
) error {
	z := zip.NewWriter(srcW)
	defer z.Close()
	putToZip := func(path string, buff []byte) (int, error) {
		w, err := z.Create(path)
		if err != nil {
			return 0, err
		}
		for i := range buff {
			buff[i] ^= key
		}
		return w.Write(buff)
	}
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
			_, err = putToZip(path, aAva)
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
			_, err = putToZip(path, aAva)
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
		// ????????????
		{
			cover, err := fetcher.ComicCover(comicIdx, comicInfo)
			if err != nil {
				return err
			}
			if cover != nil {
				path := uuid.New().String()
				_, err = putToZip(path, cover)
				if err != nil {
					return err
				}
				archive.Comics[comicIdx].CoverPath = path
			}
		}
		// ????????????
		{
			aAva, err := fetcher.ComicAuthorAvatar(comicIdx, comicInfo)
			if err != nil {
				return err
			}
			if aAva != nil {
				path := uuid.New().String()
				_, err = putToZip(path, aAva)
				if err != nil {
					return err
				}
				archive.Comics[comicIdx].AuthorAvatarPath = path
			}
		}
		// ??????
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
			// ????????????
			{
				cover, err := fetcher.VolumeCover(comicIdx, comicInfo, volumeIdx, volumesInfo)
				if err != nil {
					return err
				}
				if cover != nil {
					path := uuid.New().String()
					_, err = putToZip(path, cover)
					if err != nil {
						return err
					}
					archive.Comics[comicIdx].Volumes[volumeIdx].CoverPath = path
				}
			}
			// ??????
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
				// ??????
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
							_, err = putToZip(path, cover)
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
	_, err = putToZip(infoPath, buff)
	return err
}

func ReadPkzPath(filePath string, innerPath string) ([]byte, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	re, err := r.Open(innerPath)
	if err != nil {
		return nil, err
	}
	defer re.Close()
	buff, err := ioutil.ReadAll(re)
	if err != nil {
		return nil, err
	}
	for i := range buff {
		buff[i] ^= key
	}
	return buff, err
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
