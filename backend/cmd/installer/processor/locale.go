package processor

// Docker operations messages
const (
	MsgPullingImage                = "正在拉取镜像：%s"
	MsgImagePullCompleted          = "镜像拉取完成：%s"
	MsgImagePullFailed             = "镜像 %s 拉取失败：%v"
	MsgRemovingWorkerContainers    = "正在移除工作容器"
	MsgStoppingContainer           = "正在停止容器 %s"
	MsgRemovingContainer           = "正在移除容器 %s"
	MsgContainerRemoved            = "已移除容器 %s"
	MsgNoWorkerContainersFound     = "未找到工作容器"
	MsgWorkerContainersRemoved     = "已移除 %d 个工作容器"
	MsgRemovingImage               = "正在移除镜像：%s"
	MsgImageRemoved                = "已移除镜像 %s"
	MsgImageNotFound               = "未找到镜像 %s（可能已移除）"
	MsgWorkerImagesRemoveCompleted = "工作容器镜像已全部移除"
	MsgEnsuringDockerNetworks      = "正在检查 Docker 网络"
	MsgDockerNetworkExists         = "Docker 网络已存在：%s"
	MsgCreatingDockerNetwork       = "正在创建 Docker 网络：%s"
	MsgDockerNetworkCreated        = "Docker 网络已创建：%s"
	MsgDockerNetworkCreateFailed   = "Docker 网络 %s 创建失败：%v"
	MsgRecreatingDockerNetwork     = "正在重建 Docker 网络并补充 compose 标签：%s"
	MsgDockerNetworkRemoved        = "Docker 网络已移除：%s"
	MsgDockerNetworkRemoveFailed   = "Docker 网络 %s 移除失败：%v"
	MsgDockerNetworkInUse          = "Docker 网络 %s 正被容器使用，暂时不能重建"
)

// File system operations messages
const (
	MsgExtractingDockerCompose          = "Extracting docker-compose.yml"
	MsgExtractingLangfuseCompose        = "Extracting docker-compose-langfuse.yml"
	MsgExtractingObservabilityCompose   = "Extracting docker-compose-observability.yml"
	MsgExtractingObservabilityDirectory = "Extracting observability directory"
	MsgSkippingExternalLangfuse         = "Skipping external Langfuse deployment"
	MsgSkippingExternalObservability    = "Skipping external Observability deployment"
	MsgPatchingComposeFile              = "Patching docker-compose file: %s"
	MsgComposePatchCompleted            = "Docker-compose file patching completed"
	MsgCleaningUpStackFiles             = "Cleaning up stack files for %s"
	MsgStackFilesCleanupCompleted       = "Stack files cleanup completed"
	MsgEnsurngStackIntegrity            = "Ensuring %s stack integrity"
	MsgVerifyingStackIntegrity          = "Verifying %s stack integrity"
	MsgStackIntegrityVerified           = "Stack %s integrity verified"
	MsgUpdatingExistingFile             = "Updating existing file: %s"
	MsgCreatingMissingFile              = "Creating missing file: %s"
	MsgMissingFileNotCreated            = "Missing file %s was not created: %s"
	MsgFileIntegrityValid               = "File integrity valid: %s"
	MsgSkippingModifiedFile             = "Skipping modified files: %s"
	MsgDirectoryCheckedWithModified     = "Directory checked with modified files present: %s"
)

// Update operations messages
const (
	MsgCheckingUpdates            = "Checking for updates"
	MsgDownloadingInstaller       = "Downloading installer update"
	MsgInstallerDownloadCompleted = "Installer download completed"
	MsgUpdatingInstaller          = "Updating installer"
	MsgRemovingInstaller          = "Removing installer"
	MsgInstallerUpdateCompleted   = "Installer update completed"
	// Past tense: the length, the sha256 and the Ed25519 signature are all checked as the
	// stream arrives, so by the time this is printed the verdict is already in.
	MsgVerifyingBinaryChecksum   = "Checksum and signature verified"
	MsgInstallerSavedTo          = "Installer saved to %s"
	MsgInstallerDownloadProgress = "Downloaded %d%% (%s of %s)"
	// The installer binary outlives the removal on purpose: a process cannot delete the
	// file it is running from, and trying would leave the user with neither.
	MsgInstallerNotRemoved            = "The installer itself is left in place at %s — delete it yourself when you no longer need it"
	MsgInstallerNotRemovedUnknownPath = "The installer itself is left in place — delete it yourself when you no longer need it"
)

// Jaeger storage plugin messages. The plugin is a file mounted into the Jaeger container,
// not an image, so none of the compose output covers it.
const (
	MsgJaegerPluginDownloading       = "Downloading Jaeger storage plugin %s (%s)"
	MsgJaegerPluginProgress          = "  %s: %d%% (%s of %s)"
	MsgJaegerPluginInstalled         = "Jaeger storage plugin in place: %s"
	MsgJaegerPluginStaleRemoved      = "Removed leftover from an interrupted download: %s"
	MsgJaegerPluginStaleRemoveFailed = "Could not remove leftover %s: %v"
	MsgJaegerPluginRestarting        = "Restarting Jaeger so it loads the new storage plugin"
)

// Post-update verification messages
const (
	MsgVerifyingUpdatedComponents = "Verifying updated %s components"
	MsgComponentUpToDate          = "  %s: matches the published build"
	MsgComponentNewerThanOffered  = "  %s: newer than the offered build (registry moved ahead)"
	MsgComponentMismatch          = "  %s: expected %s, got %s"
	MsgComponentNotVerifiable     = "  %s: no published digest to compare against"
	MsgVerificationPassed         = "All updated %s components match the published build"
	MsgVerificationMismatch       = "%d %s component(s) do not match the published build"
)

// Remove operations messages
const (
	MsgRemovingStack          = "正在移除服务栈：%s"
	MsgStackRemovalCompleted  = "%s 服务栈已移除"
	MsgPurgingStack           = "正在彻底清除服务栈：%s"
	MsgStackPurgeCompleted    = "%s 服务栈已彻底清除"
	MsgExecutingDockerCompose = "正在执行 docker-compose 命令：%s"
	MsgDockerComposeCompleted = "docker-compose 命令执行完成"
	MsgFactoryResetStarting   = "正在恢复出厂设置"
	MsgFactoryResetCompleted  = "已恢复出厂设置"
	MsgRestoringDefaultEnv    = "正在从内置模板恢复默认 .env"
	MsgDefaultEnvRestored     = "默认 .env 已恢复"
)

type Subsystem string

const (
	SubsystemDocker     Subsystem = "docker"
	SubsystemCompose    Subsystem = "compose"
	SubsystemFileSystem Subsystem = "file-system"
	SubsystemUpdate     Subsystem = "update"
)

type SubsystemOperationMessage struct {
	Enter string
	Exit  string
	Error string
}

var SubsystemOperationMessages = map[Subsystem]map[ProcessorOperation]SubsystemOperationMessage{
	SubsystemCompose: {
		ProcessorOperationStart: SubsystemOperationMessage{
			Enter: "正在启动 %s compose 服务栈",
			Exit:  "%s compose 服务栈已启动",
			Error: "%s compose 服务栈启动失败",
		},
		ProcessorOperationStop: SubsystemOperationMessage{
			Enter: "正在停止 %s compose 服务栈",
			Exit:  "%s compose 服务栈已停止",
			Error: "%s compose 服务栈停止失败",
		},
		ProcessorOperationRestart: SubsystemOperationMessage{
			Enter: "正在重启 %s compose 服务栈",
			Exit:  "%s compose 服务栈已重启",
			Error: "%s compose 服务栈重启失败",
		},
		ProcessorOperationUpdate: SubsystemOperationMessage{
			Enter: "正在更新 %s compose 服务栈",
			Exit:  "%s compose 服务栈已更新",
			Error: "%s compose 服务栈更新失败",
		},
		ProcessorOperationDownload: SubsystemOperationMessage{
			Enter: "正在下载 %s compose 服务栈",
			Exit:  "%s compose 服务栈已下载",
			Error: "%s compose 服务栈下载失败",
		},
		ProcessorOperationRemove: SubsystemOperationMessage{
			Enter: "正在移除 %s compose 服务栈",
			Exit:  "%s compose 服务栈已移除",
			Error: "%s compose 服务栈移除失败",
		},
		ProcessorOperationPurge: SubsystemOperationMessage{
			Enter: "正在彻底清除 %s compose 服务栈",
			Exit:  "%s compose 服务栈已彻底清除",
			Error: "%s compose 服务栈清除失败",
		},
	},
}
