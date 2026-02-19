package env

type OS string

const (
	OSUbuntu  OS = "ubuntu"
	OSDebian  OS = "debian"
	OSFedora  OS = "fedora"
	OSArch    OS = "arch"
	OSMacOS   OS = "macos"
	OSUnknown OS = "unknown"
)

type PackageManager string

const (
	PMNone   PackageManager = "none"
	PMApt    PackageManager = "apt"
	PMDnf    PackageManager = "dnf"
	PMYum    PackageManager = "yum"
	PMPacman PackageManager = "pacman"
	PMBrew   PackageManager = "brew"
)

type Architecture string

const (
	ArchAMD64   Architecture = "amd64"
	ArchARM64   Architecture = "arm64"
	ArchUnknown Architecture = "unknown"
)

type Runtime struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Environment struct {
	OS             OS             `json:"os"`
	OSVersion      string         `json:"os_version"`
	Architecture   Architecture   `json:"architecture"`
	PackageManager PackageManager `json:"package_manager"`
	Runtimes       []Runtime      `json:"runtimes"`
	HasSudo        bool           `json:"has_sudo"`
	InContainer    bool           `json:"in_container"`
}
