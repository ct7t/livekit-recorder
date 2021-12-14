// +build !test

package pipeline

import (
	"fmt"

	livekit "github.com/livekit/protocol/proto"
	"github.com/tinyzimmer/go-gst/gst"
)

type InputBin struct {
	isStream      bool
	isPlaylist    bool
	bin           *gst.Bin
	audioElements []*gst.Element
	videoElements []*gst.Element
	audioQueue    *gst.Element
	videoQueue    *gst.Element
	mux           *gst.Element
}

func newInputBin(isStream bool, isPlaylist bool, options *livekit.RecordingOptions) (*InputBin, error) {
	// create audio elements
	pulseSrc, err := gst.NewElement("pulsesrc")
	if err != nil {
		return nil, err
	}

	audioConvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, err
	}

	audioCapsFilter, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, err
	}
	err = audioCapsFilter.SetProperty("caps", gst.NewCapsFromString(
		fmt.Sprintf("audio/x-raw,format=S16LE,layout=interleaved,rate=%d,channels=2", options.AudioFrequency),
	))
	if err != nil {
		return nil, err
	}

	faac, err := gst.NewElement("faac")
	if err != nil {
		return nil, err
	}
	err = faac.SetProperty("bitrate", int(options.AudioBitrate*1000))
	if err != nil {
		return nil, err
	}

	audioQueue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}
	if err = audioQueue.SetProperty("max-size-time", uint64(3e9)); err != nil {
		return nil, err
	}

	// create video elements
	xImageSrc, err := gst.NewElement("ximagesrc")
	if err != nil {
		return nil, err
	}
	err = xImageSrc.SetProperty("show-pointer", false)
	if err != nil {
		return nil, err
	}

	videoConvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, err
	}

	videoCapsFilter, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, err
	}
	err = videoCapsFilter.SetProperty("caps", gst.NewCapsFromString(
		fmt.Sprintf("video/x-raw,framerate=%d/1", options.Framerate),
	))
	if err != nil {
		return nil, err
	}

	x264Enc, err := gst.NewElement("x264enc")
	if err != nil {
		return nil, err
	}
	err = x264Enc.SetProperty("bitrate", uint(options.VideoBitrate))
	if err != nil {
		return nil, err
	}
	x264Enc.SetArg("speed-preset", "veryfast")
	x264Enc.SetArg("tune", "zerolatency")

	videoQueue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}
	if err = videoQueue.SetProperty("max-size-time", uint64(3e9)); err != nil {
		return nil, err
	}

	// create mux
	var mux *gst.Element
	if !isPlaylist {
		if isStream {
			mux, err = gst.NewElement("flvmux")
			if err != nil {
				return nil, err
			}
			err = mux.Set("streamable", true)
			if err != nil {
				return nil, err
			}
		} else {
			mux, err = gst.NewElement("mp4mux")
			if err != nil {
				return nil, err
			}
			err = mux.SetProperty("faststart", true)
			if err != nil {
				return nil, err
			}
		}
	}

	// create bin
	bin := gst.NewBin("input")
	err = bin.AddMany(
		// audio
		pulseSrc, audioConvert, audioCapsFilter, faac, audioQueue,
		// video
		xImageSrc, videoConvert, videoCapsFilter, x264Enc, videoQueue,
	)
	if err != nil {
		return nil, err
	}

	if isPlaylist {
		// create ghost pads - separate video and audio for playlist
		videoGhostPad := gst.NewGhostPad("video", videoQueue.GetStaticPad("src"))
		if !bin.AddPad(videoGhostPad.Pad) {
			return nil, ErrGhostPadFailed
		}
		audioGhostPad := gst.NewGhostPad("audio", audioQueue.GetStaticPad("src"))
		if !bin.AddPad(audioGhostPad.Pad) {
			return nil, ErrGhostPadFailed
		}
	} else {
		// add mux to bin if one was set
		if mux != nil {
			err = bin.AddMany(
				// mux
				mux,
			)
			if err != nil {
				return nil, err
			}
		}

		// create ghost pad - single src for non-playlist
		ghostPad := gst.NewGhostPad("src", mux.GetStaticPad("src"))
		if !bin.AddPad(ghostPad.Pad) {
			return nil, ErrGhostPadFailed
		}
	}

	return &InputBin{
		isStream:      isStream,
		isPlaylist:    isPlaylist,
		bin:           bin,
		audioElements: []*gst.Element{pulseSrc, audioConvert, audioCapsFilter, faac, audioQueue},
		videoElements: []*gst.Element{xImageSrc, videoConvert, videoCapsFilter, x264Enc, videoQueue},
		audioQueue:    audioQueue,
		videoQueue:    videoQueue,
		mux:           mux,
	}, nil
}

func (b *InputBin) Link() error {
	// link audio elements
	if err := gst.ElementLinkMany(b.audioElements...); err != nil {
		return err
	}

	// link video elements
	if err := gst.ElementLinkMany(b.videoElements...); err != nil {
		return err
	}

	// link audio and video queues to mux
	var audioPad, videoPad *gst.Pad
	if !b.isPlaylist {
		if b.isStream {
			audioPad = b.mux.GetRequestPad("audio")
			videoPad = b.mux.GetRequestPad("video")
		} else {
			audioPad = b.mux.GetRequestPad("audio_%u")
			videoPad = b.mux.GetRequestPad("video_%u")
		}
		if err := requireLink(b.audioQueue.GetStaticPad("src"), audioPad); err != nil {
			return err
		}
		if err := requireLink(b.videoQueue.GetStaticPad("src"), videoPad); err != nil {
			return err
		}
	}
	return nil
}
