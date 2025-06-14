package constant

import (
	"strings"

	"github.com/sagernet/srsc/common/semver"
)

type Platform string

const (
	PlatformUnknown Platform = ""
	PlatformSingBox Platform = "sing-box"
)

type System string

const (
	SystemUnknown   System = ""
	SystemAndroid   System = "android"
	SystemiOS       System = "ios"
	SystemMacOS     System = "macos"
	SystemAppleTVOS System = "tvos"
)

type Metadata struct {
	UserAgent string
	Platform  Platform
	System    System
	Version   *semver.Version
}

func DetectMetadata(userAgent string) Metadata {
	var metadata Metadata
	metadata.UserAgent = userAgent
	if strings.HasPrefix(userAgent, "SFA") {
		metadata.System = SystemAndroid
	} else if strings.HasPrefix(userAgent, "SFI") {
		metadata.System = SystemiOS
	} else if strings.HasPrefix(userAgent, "SFM") {
		metadata.System = SystemMacOS
	} else if strings.HasPrefix(userAgent, "SFT") {
		metadata.System = SystemAppleTVOS
	}
	var versionName string
	if strings.Contains(userAgent, "sing-box ") {
		metadata.Platform = PlatformSingBox
		versionName = strings.Split(userAgent, "sing-box ")[1]
		if strings.Contains(versionName, ";") {
			versionName = strings.Split(versionName, ";")[0]
		} else {
			versionName = strings.Split(versionName, ")")[0]
		}
	}
	if semver.IsValid(versionName) {
		version := semver.ParseVersion(versionName)
		metadata.Version = &version
	}
	return metadata
}
