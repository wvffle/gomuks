// gomuks - A terminal Matrix client written in Go.
// Copyright (C) 2020 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/pushrules"

	"maunium.net/go/gomuks/debug"
	"maunium.net/go/gomuks/matrix/rooms"
)

type AuthCache struct {
	NextBatch       string `yaml:"next_batch"`
	FilterID        string `yaml:"filter_id"`
	InitialSyncDone bool   `yaml:"initial_sync_done"`
}

type UserPreferences struct {
	HideUserList         bool `yaml:"hide_user_list"`
	HideRoomList         bool `yaml:"hide_room_list"`
	BareMessageView      bool `yaml:"bare_message_view"`
	DisableImages        bool `yaml:"disable_images"`
	DisableTypingNotifs  bool `yaml:"disable_typing_notifs"`
	DisableEmojis        bool `yaml:"disable_emojis"`
	DisableMarkdown      bool `yaml:"disable_markdown"`
	DisableHTML          bool `yaml:"disable_html"`
	DisableDownloads     bool `yaml:"disable_downloads"`
	DisableNotifications bool `yaml:"disable_notifications"`
	DisableShowURLs      bool `yaml:"disable_show_urls"`
}

type KeyMap struct {
	VerificationDone   string `yaml:"verification_done"`   // Used when verification process finished to close the modal
	VerificationSubmit string `yaml:"verification_submit"` // Used to submit "yes"/"no" to confirm/reject verification

	FuzzySearchOpen   string `yaml:"fuzzy_search_open"`   // Used to open fuzzy search modal
	FuzzySearchClose  string `yaml:"fuzzy_search_cancel"` // Used to close fuzzy search modal
	FuzzySearchNext   string `yaml:"fuzzy_search_next"`   // Used to go to next entry
	FuzzySearchPrev   string `yaml:"fuzzy_search_prev"`   // Used to go to previous entry
	FuzzySearchChoose string `yaml:"fuzzy_search_choose"` // Used to choose entry

	RoomNext string `yaml:"room_next"` // Used to go to next room
	RoomPrev string `yaml:"room_prev"` // Used to go to previous room

	RoomViewTop        string `yaml:"room_view_top"`         // Used to go to the top of the room view
	RoomViewBottom     string `yaml:"room_view_bottom"`      // Used to go to the bottom of the room view
	RoomViewScrollUp   string `yaml:"room_view_scroll_up"`   // Used to scroll room view up
	RoomViewScrollDown string `yaml:"room_view_scroll_down"` // Used to scroll room view down

	MessageSelectCancel string `yaml:"message_select_cancel"` // Used to exit message select mode
	MessageSelectNext   string `yaml:"message_select_next"`   // Used to go to next message
	MessageSelectPrev   string `yaml:"message_select_prev"`   // Used to go to previous message
	MessageSelectChoose string `yaml:"message_select_choose"` // Used to select message

	MessageInputNewline string `yaml:"message_input_newline"` // Used to insert a newline in message input
	MessageInputClear   string `yaml:"message_input_clear"`   // Used to clear input context
	MessageInputSend    string `yaml:"message_input_send"`    // Used to send message

	BareViewOpen string `yaml:"bare_view_open"` // Used to open bare messages view
}

// Config contains the main config of gomuks.
type Config struct {
	UserID      id.UserID   `yaml:"mxid"`
	DeviceID    id.DeviceID `yaml:"device_id"`
	AccessToken string      `yaml:"access_token"`
	HS          string      `yaml:"homeserver"`

	RoomCacheSize int   `yaml:"room_cache_size"`
	RoomCacheAge  int64 `yaml:"room_cache_age"`

	NotifySound        bool `yaml:"notify_sound"`
	SendToVerifiedOnly bool `yaml:"send_to_verified_only"`

	Dir          string `yaml:"-"`
	DataDir      string `yaml:"data_dir"`
	CacheDir     string `yaml:"cache_dir"`
	HistoryPath  string `yaml:"history_path"`
	RoomListPath string `yaml:"room_list_path"`
	MediaDir     string `yaml:"media_dir"`
	DownloadDir  string `yaml:"download_dir"`
	StateDir     string `yaml:"state_dir"`

	Keymap string `yaml:"keymap"`
	KeyMap KeyMap `yaml:"-"`

	Preferences UserPreferences        `yaml:"-"`
	AuthCache   AuthCache              `yaml:"-"`
	Rooms       *rooms.RoomCache       `yaml:"-"`
	PushRules   *pushrules.PushRuleset `yaml:"-"`

	nosave bool
}

// NewConfig creates a config that loads data from the given directory.
func NewConfig(configDir, dataDir, cacheDir, downloadDir string) *Config {
	return &Config{
		Dir:          configDir,
		DataDir:      dataDir,
		CacheDir:     cacheDir,
		DownloadDir:  downloadDir,
		HistoryPath:  filepath.Join(cacheDir, "history.db"),
		RoomListPath: filepath.Join(cacheDir, "rooms.gob.gz"),
		StateDir:     filepath.Join(cacheDir, "state"),
		MediaDir:     filepath.Join(cacheDir, "media"),

		Keymap: "default",

		RoomCacheSize: 32,
		RoomCacheAge:  1 * 60,

		NotifySound:        true,
		SendToVerifiedOnly: false,
	}
}

// Clear clears the session cache and removes all history.
func (config *Config) Clear() {
	_ = os.Remove(config.HistoryPath)
	_ = os.Remove(config.RoomListPath)
	_ = os.RemoveAll(config.StateDir)
	_ = os.RemoveAll(config.MediaDir)
	_ = os.RemoveAll(config.CacheDir)
	config.nosave = true
}

// ClearData clears non-temporary session data.
func (config *Config) ClearData() {
	_ = os.RemoveAll(config.DataDir)
}

func (config *Config) CreateCacheDirs() {
	_ = os.MkdirAll(config.CacheDir, 0700)
	_ = os.MkdirAll(config.DataDir, 0700)
	_ = os.MkdirAll(config.StateDir, 0700)
	_ = os.MkdirAll(config.MediaDir, 0700)
}

func (config *Config) DeleteSession() {
	config.AuthCache.NextBatch = ""
	config.AuthCache.InitialSyncDone = false
	config.AccessToken = ""
	config.DeviceID = ""
	config.Rooms = rooms.NewRoomCache(config.RoomListPath, config.StateDir, config.RoomCacheSize, config.RoomCacheAge, config.GetUserID)
	config.PushRules = nil

	config.ClearData()
	config.Clear()
	config.nosave = false
	config.CreateCacheDirs()
}

func (config *Config) LoadAll() {
	config.Load()
	config.Rooms = rooms.NewRoomCache(config.RoomListPath, config.StateDir, config.RoomCacheSize, config.RoomCacheAge, config.GetUserID)
	config.LoadAuthCache()
	config.LoadPushRules()
	config.LoadPreferences()
	err := config.Rooms.LoadList()
	if err != nil {
		panic(err)
	}
}

// Load loads the config from config.yaml in the directory given to the config struct.
func (config *Config) Load() {
	config.load("config", config.Dir, "config.yaml", config)
	config.CreateCacheDirs()
}

func (config *Config) SaveAll() {
	config.Save()
	config.SaveAuthCache()
	config.SavePushRules()
	config.SavePreferences()
	err := config.Rooms.SaveList()
	if err != nil {
		panic(err)
	}
	config.Rooms.SaveLoadedRooms()
}

// Save saves this config to config.yaml in the directory given to the config struct.
func (config *Config) Save() {
	config.save("config", config.Dir, "config.yaml", config)
}

func (config *Config) LoadPreferences() {
	config.load("user preferences", config.CacheDir, "preferences.yaml", &config.Preferences)
}

func (config *Config) SavePreferences() {
	config.save("user preferences", config.CacheDir, "preferences.yaml", &config.Preferences)
}

func (config *Config) LoadAuthCache() {
	config.load("auth cache", config.CacheDir, "auth-cache.yaml", &config.AuthCache)
}

func (config *Config) SaveAuthCache() {
	config.save("auth cache", config.CacheDir, "auth-cache.yaml", &config.AuthCache)
}

func (config *Config) LoadPushRules() {
	config.load("push rules", config.CacheDir, "pushrules.json", &config.PushRules)
}

func (config *Config) SavePushRules() {
	if config.PushRules == nil {
		return
	}
	config.save("push rules", config.CacheDir, "pushrules.json", &config.PushRules)
}

func (config *Config) LoadKeymap() {
	if config.Keymap == "default" {
		// TODO: Set default config
	}

	config.load("keymap", filepath.Join(config.Dir, "keymaps"), config.Keymap+".yaml", &config.KeyMap)
}

func (config *Config) load(name, dir, file string, target interface{}) {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		debug.Print("Failed to create", dir)
		panic(err)
	}

	path := filepath.Join(dir, file)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		debug.Print("Failed to read", name, "from", path)
		panic(err)
	}

	if strings.HasSuffix(file, ".yaml") {
		err = yaml.Unmarshal(data, target)
	} else {
		err = json.Unmarshal(data, target)
	}
	if err != nil {
		debug.Print("Failed to parse", name, "at", path)
		panic(err)
	}
}

func (config *Config) save(name, dir, file string, source interface{}) {
	if config.nosave {
		return
	}

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		debug.Print("Failed to create", dir)
		panic(err)
	}
	var data []byte
	if strings.HasSuffix(file, ".yaml") {
		data, err = yaml.Marshal(source)
	} else {
		data, err = json.Marshal(source)
	}
	if err != nil {
		debug.Print("Failed to marshal", name)
		panic(err)
	}

	path := filepath.Join(dir, file)
	err = ioutil.WriteFile(path, data, 0600)
	if err != nil {
		debug.Print("Failed to write", name, "to", path)
		panic(err)
	}
}

func (config *Config) GetUserID() id.UserID {
	return config.UserID
}

func (config *Config) SaveFilterID(_ id.UserID, filterID string) {
	config.AuthCache.FilterID = filterID
	config.SaveAuthCache()
}

func (config *Config) LoadFilterID(_ id.UserID) string {
	return config.AuthCache.FilterID
}

func (config *Config) SaveNextBatch(_ id.UserID, nextBatch string) {
	config.AuthCache.NextBatch = nextBatch
	config.SaveAuthCache()
}

func (config *Config) LoadNextBatch(_ id.UserID) string {
	return config.AuthCache.NextBatch
}

func (config *Config) SaveRoom(_ *mautrix.Room) {
	panic("SaveRoom is not supported")
}

func (config *Config) LoadRoom(_ id.RoomID) *mautrix.Room {
	panic("LoadRoom is not supported")
}
