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
	MsgExtractingDockerCompose          = "正在提取 docker-compose.yml"
	MsgExtractingLangfuseCompose        = "正在提取 docker-compose-langfuse.yml"
	MsgExtractingObservabilityCompose   = "正在提取 docker-compose-observability.yml"
	MsgExtractingObservabilityDirectory = "正在提取 observability 目录"
	MsgSkippingExternalLangfuse         = "Langfuse 使用外部服务，跳过内置部署"
	MsgSkippingExternalObservability    = "可观测性组件使用外部服务，跳过内置部署"
	MsgPatchingComposeFile              = "正在修补 docker-compose 文件：%s"
	MsgComposePatchCompleted            = "docker-compose 文件修补完成"
	MsgCleaningUpStackFiles             = "正在清理 %s 服务栈文件"
	MsgStackFilesCleanupCompleted       = "服务栈文件清理完成"
	MsgEnsurngStackIntegrity            = "正在校正 %s 服务栈文件"
	MsgVerifyingStackIntegrity          = "正在验证 %s 服务栈完整性"
	MsgStackIntegrityVerified           = "%s 服务栈完整性验证通过"
	MsgUpdatingExistingFile             = "正在更新现有文件：%s"
	MsgCreatingMissingFile              = "正在创建缺失文件：%s"
	MsgFileIntegrityValid               = "文件完整性验证通过：%s"
	MsgSkippingModifiedFile             = "文件已被修改，跳过更新：%s"
	MsgDirectoryCheckedWithModified     = "目录检查完成，其中包含已修改文件：%s"
)

// Update operations messages
const (
	MsgCheckingUpdates            = "正在检查更新"
	MsgDownloadingInstaller       = "正在下载安装程序更新"
	MsgInstallerDownloadCompleted = "安装程序更新下载完成"
	MsgUpdatingInstaller          = "正在更新安装程序"
	MsgRemovingInstaller          = "正在移除安装程序"
	MsgInstallerUpdateCompleted   = "安装程序更新完成"
	MsgVerifyingBinaryChecksum    = "正在验证二进制文件校验和"
	MsgReplacingInstallerBinary   = "正在替换安装程序二进制文件"
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
