package query

// Platform is where the game is played (i.e. console, operating system)
type Platform string

// SortBy is how the response will be sorted
type SortBy string

// LengthRange allows filtering by the time it takes to complete a game
type LengthRange string

// Modifier is additional information to be included in response
type Modifier string

// SortDirection specifies the direction the responses will be sorted (baed on SortBy)
type SortDirection string

// QueryType is the subject being queried
type QueryType string

const (
	// Platforms

	ThreeDO              Platform = "3DO"
	Amiga                Platform = "Amiga"
	AmstradCPC           Platform = "Amstrad CPC"
	Android              Platform = "Android"
	AppleII              Platform = "Apple II"
	Arcade               Platform = "Arcade"
	Atari2600            Platform = "Atari 2600"
	Atari5200            Platform = "Atari 5200"
	Atari7800            Platform = "Atari 7800"
	Atari8bitFamily      Platform = "Atari 8-bit Family"
	AtariJaguar          Platform = "Atari Jaguar"
	AtariJaguarCD        Platform = "Atari Jaguar CD"
	AtariLynx            Platform = "Atari Lynx"
	AtariST              Platform = "Atari ST"
	BBCMicro             Platform = "BBC Micro"
	Browser              Platform = "Browser"
	ColecoVision         Platform = "ColecoVision"
	Commodore64          Platform = "Commodore 64"
	Dreamcast            Platform = "Dreamcast"
	Emulated             Platform = "Emulated"
	FMTowns              Platform = "FM Towns"
	GameWatch            Platform = "Game & Watch"
	GameBoy              Platform = "Game Boy"
	GameBoyAdvance       Platform = "Game Boy Advance"
	GameBoyColor         Platform = "Game Boy Color"
	GearVR               Platform = "Gear VR"
	GoogleStadia         Platform = "Google Stadia"
	Intellivision        Platform = "Intellivision"
	InteractiveMovie     Platform = "Interactive Movie"
	iOS                  Platform = "iOS"
	Linux                Platform = "Linux"
	Mac                  Platform = "Mac"
	Mobile               Platform = "Mobile"
	MSX                  Platform = "MSX"
	NGage                Platform = "N-Gage"
	NECPC8800            Platform = "NEC PC-8800"
	NECPC980121          Platform = "NEC PC-9801/21"
	NECPCFX              Platform = "NEC PC-FX"
	NeoGeo               Platform = "Neo Geo"
	NeoGeoCD             Platform = "Neo Geo CD"
	NeoGeoPocket         Platform = "Neo Geo Pocket"
	NES                  Platform = "NES"
	Nintendo3DS          Platform = "Nintendo 3DS"
	Nintendo64           Platform = "Nintendo 64"
	NintendoDS           Platform = "Nintendo DS"
	NintendoGameCube     Platform = "Nintendo GameCube"
	NintendoSwitch       Platform = "Nintendo Switch"
	OculusGo             Platform = "Oculus Go"
	OculusQuest          Platform = "Oculus Quest"
	OnLive               Platform = "OnLive"
	Ouya                 Platform = "Ouya"
	PC                   Platform = "PC"
	PCVR                 Platform = "PC VR"
	PhilipsCDi           Platform = "Philips CD-i"
	PhilipsVideopacG7000 Platform = "Philips Videopac G7000"
	PlayStation          Platform = "PlayStation"
	PlayStation2         Platform = "PlayStation 2"
	PlayStation3         Platform = "PlayStation 3"
	PlayStation4         Platform = "PlayStation 4"
	PlayStation5         Platform = "PlayStation 5"
	PlayStationMobile    Platform = "PlayStation Mobile"
	PlayStationNow       Platform = "PlayStation Now"
	PlayStationPortable  Platform = "PlayStation Portable"
	PlayStationVita      Platform = "PlayStation Vita"
	PlayStationVR        Platform = "PlayStation VR"
	PlugPlay             Platform = "Plug & Play"
	Sega32X              Platform = "Sega 32X"
	SegaCD               Platform = "Sega CD"
	SegaGameGear         Platform = "Sega Game Gear"
	SegaMasterSystem     Platform = "Sega Master System"
	SegaMegaDriveGenesis Platform = "Sega Mega Drive/Genesis"
	SegaSaturn           Platform = "Sega Saturn"
	SG1000               Platform = "SG-1000"
	SharpX68000          Platform = "Sharp X68000"
	SuperNintendo        Platform = "Super Nintendo"
	TigerHandheld        Platform = "Tiger Handheld"
	TurboGrafx16         Platform = "TurboGrafx-16"
	TurboGrafxCD         Platform = "TurboGrafx-CD"
	VirtualBoy           Platform = "Virtual Boy"
	Wii                  Platform = "Wii"
	WiiU                 Platform = "Wii U"
	WindowsPhone         Platform = "Windows Phone"
	WonderSwan           Platform = "WonderSwan"
	Xbox                 Platform = "Xbox"
	Xbox360              Platform = "Xbox 360"
	XboxOne              Platform = "Xbox One"
	XboxSeriesXS         Platform = "Xbox Series X/S"
	ZXSpectrum           Platform = "ZX Spectrum"

	/*
		Sort by

		Specifies how the responses should be sorted.
		Responses will be sorted by game title by default.
	*/

	// Name sorts by game title - default
	Name SortBy = "name"
	// MainStory sorts by the main story competion time (shortest to longest)
	MainStory SortBy = "main"
	// MainExtras sorts by the main + extras competion time (shortest to longest)
	MainExtras SortBy = "mainp"
	// Completionist sorts by the completionist competion time (shortest to longest)
	Completionist SortBy = "comp"
	// AverageTime sorts by the average competion time (shortest to longest)
	AverageTime SortBy = "averagea"
	// TopRated sorts by games with the highest user rating (highest to lowest)
	TopRated SortBy = "rating"
	// MostPopular sorts by games that have been added by the most number of users
	MostPopular SortBy = "popular"
	// MostBacklogs sorts by the number of users with the game in their backlog
	MostBacklogs SortBy = "backlog"
	// MostSubmissions sorts by the number of user submissions (submitted a time)
	MostSubmissions SortBy = "usersp"
	// MostPlayed sorts by the number of users that have completed the game
	MostPlayed SortBy = "playing"
	// MostSpeedruns sorts by the number of submitted speedruns for a game
	// (note that these are the speed runs submitted to howlongtobeat.com, not speedrun.com)
	MostSpeedruns SortBy = "speedruns"
	// ReleaseDate sorts by the release date of the game (earliest to most recent)
	ReleaseDate SortBy = "release"

	// Length Range - used when search by time range

	// RangeMainStory will search Main Story completion times
	RangeMainStory LengthRange = "main"
	// RangeMainExtras will search Main + Extra completion times
	RangeMainExtras LengthRange = "mainp"
	// RangeCompletionist will search Completionist completion times
	RangeCompletionist LengthRange = "comp"
	// RangeAverageTime will search average completion times
	RangeAverageTime LengthRange = "averagea"

	/*
		Modifiers

		Modifiers offer additional options to return additional information in a query.
		Most of the time you will not be using these
	*/

	// NoModifier will use no search modifiers
	NoModifier Modifier = ""
	// IncludeDLC will show DLC in responses
	IncludeDLC Modifier = "show_dlc"
	// IsolateDLC will only return DCL in responses
	IsolateDLC Modifier = "only_dlc"
	// HiddenStats hides all responses - this is pointless for this tool. Leaving for documentation purposes.
	// HiddenStats Modifier = "hidden_stats"

	// ShowUserStats will include user stats in responses
	ShowUserStats Modifier = "user_stats"

	/*
		Sort Direction

		Specifies the direction that responses will be sorted.
		Default direction determined by SortBy method used.
	*/

	// NormalOrder sorts in the direction specified by the SortBy method
	NormalOrder SortDirection = "Normal Order"
	// ReverseOrder sorts in the reverse direction specified by the SortBy method
	ReverseOrder SortDirection = "Reverse Order"

	/*
		QueryType

		Specifies the kind of query that will be made. Determines
		which resource the query is running against
	*/

	// GameQuery will query against game titles
	GameQuery QueryType = "games"
	// UserQuery will query against user names
	UserQuery QueryType = "users"
)
