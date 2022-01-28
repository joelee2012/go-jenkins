package jenkins

type Job struct {
	Class                 string         `json:"_class"`
	Actions               []Actions      `json:"actions"`
	Description           string         `json:"description"`
	DisplayName           string         `json:"displayName"`
	DisplayNameOrNull     interface{}    `json:"displayNameOrNull"`
	FullDisplayName       string         `json:"fullDisplayName"`
	FullName              string         `json:"fullName"`
	Name                  string         `json:"name"`
	URL                   string         `json:"url"`
	Buildable             bool           `json:"buildable"`
	Builds                []*Build       `json:"builds"`
	Color                 string         `json:"color"`
	FirstBuild            *Build         `json:"firstBuild"`
	HealthReport          []HealthReport `json:"healthReport"`
	InQueue               bool           `json:"inQueue"`
	KeepDependencies      bool           `json:"keepDependencies"`
	LastBuild             *Build         `json:"lastBuild"`
	LastCompletedBuild    *Build         `json:"lastCompletedBuild"`
	LastFailedBuild       *Build         `json:"lastFailedBuild"`
	LastStableBuild       *Build         `json:"lastStableBuild"`
	LastSuccessfulBuild   *Build         `json:"lastSuccessfulBuild"`
	LastUnstableBuild     *Build         `json:"lastUnstableBuild"`
	LastUnsuccessfulBuild *Build         `json:"lastUnsuccessfulBuild"`
	NextBuildNumber       int            `json:"nextBuildNumber"`
	Property              []Property     `json:"property"`
	QueueItem             interface{}    `json:"queueItem"`
	ConcurrentBuild       bool           `json:"concurrentBuild"`
	ResumeBlocked         bool           `json:"resumeBlocked"`
	Jobs                  []*Job         `json:"jobs"`
	PrimaryView           *PrimaryView   `json:"primaryView"`
	Views                 []*Views       `json:"views"`
}

type Build struct {
	Class             string       `json:"_class"`
	Actions           []Actions    `json:"actions"`
	Artifacts         []Artifacts  `json:"artifacts"`
	Building          bool         `json:"building"`
	Description       interface{}  `json:"description"`
	DisplayName       string       `json:"displayName"`
	Duration          int          `json:"duration"`
	EstimatedDuration int          `json:"estimatedDuration"`
	Executor          interface{}  `json:"executor"`
	FullDisplayName   string       `json:"fullDisplayName"`
	ID                string       `json:"id"`
	KeepLog           bool         `json:"keepLog"`
	Number            int          `json:"number"`
	QueueID           int          `json:"queueId"`
	Result            string       `json:"result"`
	Timestamp         int64        `json:"timestamp"`
	URL               string       `json:"url"`
	ChangeSets        []ChangeSets `json:"changeSets"`
	Culprits          []Culprits   `json:"culprits"`
	NextBuild         *Build       `json:"nextBuild"`
	PreviousBuild     *Build       `json:"previousBuild"`
}
type Parameters struct {
	Class string `json:"_class"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
type Causes struct {
	Class            string `json:"_class"`
	ShortDescription string `json:"shortDescription"`
	UserID           string `json:"userId"`
	UserName         string `json:"userName"`
}
type Branch struct {
	SHA1 string `json:"SHA1"`
	Name string `json:"name"`
}
type Marked struct {
	SHA1   string   `json:"SHA1"`
	Branch []Branch `json:"branch"`
}
type Revision struct {
	SHA1   string   `json:"SHA1"`
	Branch []Branch `json:"branch"`
}
type RefsRemotesOriginDev struct {
	Class       string      `json:"_class"`
	BuildNumber int         `json:"buildNumber"`
	BuildResult interface{} `json:"buildResult"`
	Marked      Marked      `json:"marked"`
	Revision    Revision    `json:"revision"`
}

type LastBuiltRevision struct {
	SHA1   string   `json:"SHA1"`
	Branch []Branch `json:"branch"`
}
type RefsRemotesOriginDevFramework struct {
	Class       string      `json:"_class"`
	BuildNumber int         `json:"buildNumber"`
	BuildResult interface{} `json:"buildResult"`
	Marked      Marked      `json:"marked"`
	Revision    Revision    `json:"revision"`
}
type RefsRemotesOriginDevHush struct {
	Class       string      `json:"_class"`
	BuildNumber int         `json:"buildNumber"`
	BuildResult interface{} `json:"buildResult"`
	Marked      Marked      `json:"marked"`
	Revision    Revision    `json:"revision"`
}
type BuildsByBranchName struct {
	RefsRemotesOriginDevFramework RefsRemotesOriginDevFramework `json:"refs/remotes/origin/dev_framework"`
	RefsRemotesOriginDevHush      RefsRemotesOriginDevHush      `json:"refs/remotes/origin/dev_hush"`
}
type RefsRemotesOriginMPlayer struct {
	Class       string      `json:"_class"`
	BuildNumber int         `json:"buildNumber"`
	BuildResult interface{} `json:"buildResult"`
	Marked      Marked      `json:"marked"`
	Revision    Revision    `json:"revision"`
}

type Actions struct {
	Class                   string             `json:"_class,omitempty"`
	Parameters              []Parameters       `json:"parameters,omitempty"`
	Causes                  []Causes           `json:"causes,omitempty"`
	BlockedDurationMillis   int                `json:"blockedDurationMillis,omitempty"`
	BlockedTimeMillis       int                `json:"blockedTimeMillis,omitempty"`
	BuildableDurationMillis int                `json:"buildableDurationMillis,omitempty"`
	BuildableTimeMillis     int                `json:"buildableTimeMillis,omitempty"`
	BuildingDurationMillis  int                `json:"buildingDurationMillis,omitempty"`
	ExecutingTimeMillis     int                `json:"executingTimeMillis,omitempty"`
	ExecutorUtilization     float64            `json:"executorUtilization,omitempty"`
	SubTaskCount            int                `json:"subTaskCount,omitempty"`
	WaitingDurationMillis   int                `json:"waitingDurationMillis,omitempty"`
	WaitingTimeMillis       int                `json:"waitingTimeMillis,omitempty"`
	BuildsByBranchName      BuildsByBranchName `json:"buildsByBranchName,omitempty"`
	LastBuiltRevision       LastBuiltRevision  `json:"lastBuiltRevision,omitempty"`
	RemoteUrls              []string           `json:"remoteUrls,omitempty"`
	ScmName                 string             `json:"scmName,omitempty"`
}

type Artifacts struct {
	DisplayPath  string `json:"displayPath"`
	FileName     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
}
type Author struct {
	AbsoluteURL string `json:"absoluteUrl"`
	FullName    string `json:"fullName"`
}
type Paths struct {
	EditType string `json:"editType"`
	File     string `json:"file"`
}
type Items struct {
	Class         string   `json:"_class"`
	AffectedPaths []string `json:"affectedPaths"`
	CommitID      string   `json:"commitId"`
	Timestamp     int64    `json:"timestamp"`
	Author        Author   `json:"author"`
	AuthorEmail   string   `json:"authorEmail"`
	Comment       string   `json:"comment"`
	Date          string   `json:"date"`
	ID            string   `json:"id"`
	Msg           string   `json:"msg"`
	Paths         []Paths  `json:"paths"`
}
type ChangeSets struct {
	Class string  `json:"_class"`
	Items []Items `json:"items"`
	Kind  string  `json:"kind"`
}
type Culprits struct {
	AbsoluteURL string `json:"absoluteUrl"`
	FullName    string `json:"fullName"`
}

type PrimaryView struct {
	Class string `json:"_class"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}
type Views struct {
	Class string `json:"_class"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

type HealthReport struct {
	Description   string `json:"description"`
	IconClassName string `json:"iconClassName"`
	IconURL       string `json:"iconUrl"`
	Score         int    `json:"score"`
}

type DefaultParameterValue struct {
	Class string `json:"_class"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
type ParameterDefinitions struct {
	Class                 string                `json:"_class"`
	DefaultParameterValue DefaultParameterValue `json:"defaultParameterValue"`
	Description           string                `json:"description"`
	Name                  string                `json:"name"`
	Type                  string                `json:"type"`
	Choices               []string              `json:"choices"`
}
type Property struct {
	Class                string                 `json:"_class"`
	ParameterDefinitions []ParameterDefinitions `json:"parameterDefinitions,omitempty"`
}

type ComputerSet struct {
	Class          string      `json:"_class"`
	BusyExecutors  int         `json:"busyExecutors"`
	Computers      []*Computer `json:"computer"`
	DisplayName    string      `json:"displayName"`
	TotalExecutors int         `json:"totalExecutors"`
}

type Nodes struct {
	Class           string           `json:"_class"`
	AssignedLabels  []AssignedLabels `json:"assignedLabels"`
	Mode            string           `json:"mode"`
	NodeDescription string           `json:"nodeDescription"`
	NodeName        string           `json:"nodeName"`
	NumExecutors    int              `json:"numExecutors"`
	Description     interface{}      `json:"description"`
	Jobs            []interface{}    `json:"jobs"`
	PrimaryView     PrimaryView      `json:"primaryView"`
	QuietingDown    bool             `json:"quietingDown"`
	SlaveAgentPort  int              `json:"slaveAgentPort"`
	URL             string           `json:"url"`
	UseCrumbs       bool             `json:"useCrumbs"`
	UseSecurity     bool             `json:"useSecurity"`
	Views           []Views          `json:"views"`
}

type AssignedLabels struct {
	Actions        []Actions     `json:"actions"`
	BusyExecutors  int           `json:"busyExecutors"`
	Clouds         []interface{} `json:"clouds"`
	Description    interface{}   `json:"description"`
	IdleExecutors  int           `json:"idleExecutors"`
	Name           string        `json:"name"`
	Nodes          []Nodes       `json:"nodes"`
	Offline        bool          `json:"offline"`
	TiedJobs       []interface{} `json:"tiedJobs"`
	TotalExecutors int           `json:"totalExecutors"`
	PropertiesList []interface{} `json:"propertiesList"`
}

type ChangeSet struct {
	Class string        `json:"_class"`
	Items []interface{} `json:"items"`
	Kind  interface{}   `json:"kind"`
}

type Executor struct {
	Idle              bool               `json:"idle"`
	LikelyStuck       bool               `json:"likelyStuck"`
	Number            int                `json:"number"`
	Progress          int                `json:"progress"`
	CurrentExecutable *CurrentExecutable `json:"currentExecutable,omitempty"`
}

type HudsonNodeMonitorsSwapSpaceMonitor struct {
	Class                   string `json:"_class"`
	AvailablePhysicalMemory int    `json:"availablePhysicalMemory"`
	AvailableSwapSpace      int    `json:"availableSwapSpace"`
	TotalPhysicalMemory     int    `json:"totalPhysicalMemory"`
	TotalSwapSpace          int    `json:"totalSwapSpace"`
}

type HudsonNodeMonitorsTemporarySpaceMonitor struct {
	Class     string `json:"_class"`
	Timestamp int64  `json:"timestamp"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
}

type HudsonNodeMonitorsDiskSpaceMonitor struct {
	Class     string `json:"_class"`
	Timestamp int64  `json:"timestamp"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
}

type HudsonNodeMonitorsResponseTimeMonitor struct {
	Class     string `json:"_class"`
	Timestamp int64  `json:"timestamp"`
	Average   int    `json:"average"`
}

type HudsonNodeMonitorsClockMonitor struct {
	Class string `json:"_class"`
	Diff  int    `json:"diff"`
}

type MonitorData struct {
	HudsonNodeMonitorsSwapSpaceMonitor      HudsonNodeMonitorsSwapSpaceMonitor      `json:"hudson.node_monitors.SwapSpaceMonitor"`
	HudsonNodeMonitorsTemporarySpaceMonitor HudsonNodeMonitorsTemporarySpaceMonitor `json:"hudson.node_monitors.TemporarySpaceMonitor"`
	HudsonNodeMonitorsDiskSpaceMonitor      HudsonNodeMonitorsDiskSpaceMonitor      `json:"hudson.node_monitors.DiskSpaceMonitor"`
	HudsonNodeMonitorsArchitectureMonitor   string                                  `json:"hudson.node_monitors.ArchitectureMonitor"`
	HudsonNodeMonitorsResponseTimeMonitor   HudsonNodeMonitorsResponseTimeMonitor   `json:"hudson.node_monitors.ResponseTimeMonitor"`
	HudsonNodeMonitorsClockMonitor          HudsonNodeMonitorsClockMonitor          `json:"hudson.node_monitors.ClockMonitor"`
}

type PreviousBuild struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

type CurrentExecutable struct {
	Class             string        `json:"_class"`
	Actions           []Actions     `json:"actions"`
	Artifacts         []interface{} `json:"artifacts"`
	Building          bool          `json:"building"`
	Description       interface{}   `json:"description"`
	DisplayName       string        `json:"displayName"`
	Duration          int           `json:"duration"`
	EstimatedDuration int           `json:"estimatedDuration"`
	FullDisplayName   string        `json:"fullDisplayName"`
	ID                string        `json:"id"`
	KeepLog           bool          `json:"keepLog"`
	Number            int           `json:"number"`
	QueueID           int           `json:"queueId"`
	Result            interface{}   `json:"result"`
	Timestamp         int64         `json:"timestamp"`
	URL               string        `json:"url"`
	ChangeSets        []interface{} `json:"changeSets"`
	Culprits          []interface{} `json:"culprits"`
	NextBuild         interface{}   `json:"nextBuild"`
	PreviousBuild     PreviousBuild `json:"previousBuild"`
}

type OneOffExecutor struct {
	CurrentExecutable CurrentExecutable `json:"currentExecutable"`
	Idle              bool              `json:"idle"`
	LikelyStuck       bool              `json:"likelyStuck"`
	Number            int               `json:"number"`
	Progress          int               `json:"progress"`
}

type Computer struct {
	Class               string           `json:"_class"`
	Actions             []Actions        `json:"actions"`
	AssignedLabels      []AssignedLabels `json:"assignedLabels"`
	Description         string           `json:"description"`
	DisplayName         string           `json:"displayName"`
	Executors           []*Executor      `json:"executors"`
	Icon                string           `json:"icon"`
	IconClassName       string           `json:"iconClassName"`
	Idle                bool             `json:"idle"`
	JnlpAgent           bool             `json:"jnlpAgent"`
	LaunchSupported     bool             `json:"launchSupported"`
	ManualLaunchAllowed bool             `json:"manualLaunchAllowed"`
	MonitorData         MonitorData      `json:"monitorData"`
	NumExecutors        int              `json:"numExecutors"`
	Offline             bool             `json:"offline"`
	OfflineCause        interface{}      `json:"offlineCause"`
	OfflineCauseReason  string           `json:"offlineCauseReason"`
	OneOffExecutors     []*Executor      `json:"oneOffExecutors"`
	TemporarilyOffline  bool             `json:"temporarilyOffline"`
	AbsoluteRemotePath  interface{}      `json:"absoluteRemotePath,omitempty"`
}

type QueueJson struct {
	Class             string          `json:"_class"`
	DiscoverableItems []interface{}   `json:"discoverableItems"`
	Items             []QueueItemJson `json:"items"`
}

type QueueItemJson struct {
	Class        string      `json:"_class"`
	Actions      []Actions   `json:"actions"`
	Blocked      bool        `json:"blocked"`
	Buildable    bool        `json:"buildable"`
	ID           int         `json:"id"`
	InQueueSince int64       `json:"inQueueSince"`
	Params       string      `json:"params"`
	Stuck        bool        `json:"stuck"`
	Task         Task        `json:"task"`
	URL          string      `json:"url"`
	Why          interface{} `json:"why"`
	Cancelled    bool        `json:"cancelled"`
	Executable   Executable  `json:"executable"`
}

type Task struct {
	Class string `json:"_class"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

type Executable struct {
	Class  string `json:"_class"`
	Number int    `json:"number"`
	URL    string `json:"url"`
}
