import io
import json
import re

PATH = 'src/lib/i18n/locales/zh-CN.ts'

# English source -> Chinese. Brand names (PentAGI, MiniMax) stay in English,
# matching the backend policy of not translating product or vendor names.
PAIRS = [
    (", and", "、以及"),
    ("A new version was likely just deployed. Reloading will load the latest one.",
     "很可能刚刚部署了新版本，重新加载即可载入最新版。"),
    ("Account", "账户"),
    ("Align column", "列对齐"),
    ("Alt text (optional)", "替代文字（可选）"),
    ("Apply image", "应用图片"),
    ("Apply link", "应用链接"),
    ("Attach resources", "附加资源"),
    ("Blockquote", "引用块"),
    ("Bold", "加粗"),
    ("Bullet list", "无序列表"),
    ("Center", "居中"),
    ("Change", "修改"),
    ("Change your account password.", "修改账户密码。"),
    ("Clear agent search", "清空智能体搜索"),
    ("Clear contents", "清空内容"),
    ("Clear file search", "清空文件搜索"),
    ("Clear formatting", "清除格式"),
    ("Clear message search", "清空消息搜索"),
    ("Clear resource search", "清空资源搜索"),
    ("Clear screenshot search", "清空截图搜索"),
    ("Clear search", "清空搜索"),
    ("Clear task search", "清空任务搜索"),
    ("Clear terminal search", "清空终端搜索"),
    ("Clear tool search", "清空工具搜索"),
    ("Clear vector store search", "清空向量库搜索"),
    ("Code block", "代码块"),
    ("Column actions", "列操作"),
    ("Couldn't load", "加载失败"),
    ("Current Password", "当前密码"),
    ("Current password is required", "请填写当前密码"),
    ("Delete", "删除"),
    ("Delete column", "删除列"),
    ("Delete row", "删除行"),
    ("Delete table", "删除表格"),
    ("Describe the image", "描述这张图片"),
    ("Directory truncated", "目录已截断"),
    ("Display name", "显示名称"),
    ("Displays the mobile sidebar.", "显示移动端侧边栏。"),
    ("Email address", "邮箱地址"),
    ("Email is required", "请填写邮箱"),
    ("Email must not exceed 50 characters", "邮箱长度不能超过 50 个字符"),
    ("Email successfully updated", "邮箱已更新"),
    ("Enter your current password", "请输入当前密码"),
    ("Enter your display name", "请输入显示名称"),
    ("Enter your new email address", "请输入新的邮箱地址"),
    ("Failed to rename knowledge document", "重命名知识文档失败"),
    ("Formatting", "格式"),
    ("Header row", "表头行"),
    ("Heading 1", "一级标题"),
    ("Heading 2", "二级标题"),
    ("Heading 3", "三级标题"),
    ("Heading 4", "四级标题"),
    ("Heading 5", "五级标题"),
    ("Heading 6", "六级标题"),
    ("Hide password", "隐藏密码"),
    ("Horizontal rule", "分隔线"),
    ("Image URL", "图片地址"),
    ("Inline code", "行内代码"),
    ("Insert above", "在上方插入"),
    ("Insert below", "在下方插入"),
    ("Insert column left", "在左侧插入列"),
    ("Insert column right", "在右侧插入列"),
    ("Insert image", "插入图片"),
    ("Insert left", "在左侧插入"),
    ("Insert right", "在右侧插入"),
    ("Insert row above", "在上方插入行"),
    ("Insert row below", "在下方插入行"),
    ("Insert table", "插入表格"),
    ("Invalid email address", "邮箱地址无效"),
    ("Italic", "斜体"),
    ("Link", "链接"),
    ("Link URL", "链接地址"),
    ("Linked from your", "已关联你的"),
    ("Lists", "列表"),
    ("Local account", "本地账户"),
    ("Member since", "注册时间"),
    ("Name is required", "请填写名称"),
    ("Name must not exceed 70 characters", "名称长度不能超过 70 个字符"),
    ("Name successfully updated", "名称已更新"),
    ("New Email", "新邮箱"),
    ("None of the", "没有任何"),
    ("Nothing readable here", "此处没有可读内容"),
    ("OAuth account", "OAuth 账户"),
    ("Only http(s) or base64 raster image URLs are allowed.",
     "仅允许 http(s) 或 base64 位图图片地址。"),
    ("Only http, https, mailto and tel links are allowed.",
     "仅允许 http、https、mailto 和 tel 链接。"),
    ("Open link in new tab", "在新标签页打开链接"),
    ("Ordered list", "有序列表"),
    ("Ordered list item", "有序列表项"),
    ("Parent directory", "上级目录"),
    ("Password", "密码"),
    ("Pull from container", "从容器拉取"),
    ("Raw source", "源码"),
    ("Redo", "重做"),
    ("Refresh listing", "刷新列表"),
    ("Reload", "重新加载"),
    ("Remove image", "移除图片"),
    ("Remove link", "移除链接"),
    ("Rich editor", "富文本编辑器"),
    ("Right", "右对齐"),
    ("Row actions", "行操作"),
    ("Rows per page", "每页行数"),
    ("Scroll to latest agent log", "滚动到最新智能体日志"),
    ("Scroll to latest message", "滚动到最新消息"),
    ("Scroll to latest screenshot", "滚动到最新截图"),
    ("Scroll to latest task", "滚动到最新任务"),
    ("Scroll to latest tool log", "滚动到最新工具日志"),
    ("Scroll to latest vector store log", "滚动到最新向量库日志"),
    ("Select assistant", "选择助手"),
    ("Show password", "显示密码"),
    ("Sidebar", "侧边栏"),
    ("Something went wrong", "出现了问题"),
    ("Strikethrough", "删除线"),
    ("Table", "表格"),
    ("Task list", "任务列表"),
    ("Text", "正文"),
    ("Text style", "文本样式"),
    ("Text style:", "文本样式："),
    ("The email you use to sign in.", "你登录时使用的邮箱。"),
    ("The name shown across the app.", "在应用中展示的名称。"),
    ("The page hit a display glitch. Reloading usually clears it.",
     "页面出现了显示异常，重新加载通常即可恢复。"),
    ("The page ran into an unexpected error. Reloading usually clears it.",
     "页面遇到未预期的错误，重新加载通常即可恢复。"),
    ("The readable entries are shown below. Skipped:", "以下是可读的条目。已跳过："),
    ("This directory has too many entries to list in full; only the first",
     "该目录条目过多，无法完整列出；仅显示前"),
    ("Try again", "重试"),
    ("Undo", "撤销"),
    ("Update Email", "更新邮箱"),
    ("Update Name", "更新名称"),
    ("Upload files", "上传文件"),
    ("are shown. Open a subfolder to see the rest.", "条。打开子目录可查看其余内容。"),
    ("could be read.", "可被读取。"),
    ("could not be read", "无法读取"),
    ("entries", "个条目"),
    ("entry", "个条目"),
    ("more", "更多"),
    ("…and", "…以及"),
]


def quote(value):
    """Render a TS single-quoted literal, escaping via JSON then swapping quotes."""
    return "'" + json.dumps(value, ensure_ascii=False)[1:-1].replace("'", "\\'") + "'"


def main():
    src = io.open(PATH, encoding='utf-8').read()
    existing = set(re.findall(r"^\s*'((?:[^'\\]|\\.)*)':", src, re.M))

    lines = []
    for english, chinese in PAIRS:
        key = quote(english)
        if key[1:-1] in existing:
            continue
        lines.append('  %s: %s,' % (key, quote(chinese)))

    if not lines:
        print('ADDED: 0')
        return

    block = ('\n  // Markdown editor, account settings, and shared UI (upstream sync)\n'
             + '\n'.join(lines) + '\n')

    idx = src.rstrip().rfind('};')
    src = src[:idx] + block + src[idx:]
    io.open(PATH, 'w', encoding='utf-8', newline='\n').write(src)
    print('ADDED: %d' % len(lines))


main()
