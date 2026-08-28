import { readFile, readdir, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import ts from 'typescript';

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const frontendDirectory = path.resolve(scriptDirectory, '..');
const repositoryDirectory = path.resolve(frontendDirectory, '..');
const sourceDirectory = path.join(frontendDirectory, 'src');
const baselinePath = path.join(scriptDirectory, 'ui-english-baseline.txt');
const updateBaseline = process.argv.includes('--update');

const visibleAttributes = new Set([
    'alt',
    'aria-label',
    'cancelText',
    'confirmText',
    'description',
    'filterPlaceholder',
    'label',
    'placeholder',
    'sheetTitle',
    'title',
    'tooltip',
]);
const visibleProperties = new Set([
    'alt',
    'description',
    'errorMessage',
    'label',
    'message',
    'placeholder',
    'title',
    'tooltip',
]);
const visibleVariables = new Set([
    'description',
    'emptyText',
    'errorMessage',
    'label',
    'message',
    'placeholder',
    'title',
]);
const visibleVariableName = (name) =>
    visibleVariables.has(name) || /(description|error|label|message|placeholder|text|title|tooltip)$/i.test(name);

const normalize = (value) => value.replaceAll(/\s+/g, ' ').trim();
const containsEnglish = (value) => /[A-Za-z]{2,}/.test(value);
const containsChinese = (value) => /[\u3400-\u9fff]/.test(value);

const collectFiles = async (directory) => {
    const entries = await readdir(directory, { withFileTypes: true });
    const files = [];

    for (const entry of entries) {
        const fullPath = path.join(directory, entry.name);

        if (entry.isDirectory()) {
            files.push(...(await collectFiles(fullPath)));
        } else if (
            (entry.name.endsWith('.ts') || entry.name.endsWith('.tsx')) &&
            !entry.name.endsWith('.d.ts') &&
            !entry.name.includes('.test.') &&
            !entry.name.includes('.spec.')
        ) {
            files.push(fullPath);
        }
    }

    return files;
};

const expressionStrings = (node) => {
    if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) {
        return [node.text];
    }

    if (ts.isTemplateExpression(node)) {
        return [node.head.text, ...node.templateSpans.map((span) => span.literal.text)];
    }

    if (ts.isConditionalExpression(node)) {
        return [...expressionStrings(node.whenTrue), ...expressionStrings(node.whenFalse)];
    }

    if (
        ts.isBinaryExpression(node) &&
        [ts.SyntaxKind.BarBarToken, ts.SyntaxKind.PlusToken, ts.SyntaxKind.QuestionQuestionToken].includes(
            node.operatorToken.kind,
        )
    ) {
        return [...expressionStrings(node.left), ...expressionStrings(node.right)];
    }

    if (
        ts.isAsExpression(node) ||
        ts.isNonNullExpression(node) ||
        ts.isParenthesizedExpression(node) ||
        ts.isSatisfiesExpression(node)
    ) {
        return expressionStrings(node.expression);
    }

    return [];
};

const propertyName = (node) => {
    if (!node) {
        return undefined;
    }

    if (ts.isIdentifier(node) || ts.isStringLiteral(node)) {
        return node.text;
    }

    return undefined;
};

const scanFile = async (filePath) => {
    const sourceText = await readFile(filePath, 'utf8');
    const scriptKind = filePath.endsWith('.tsx') ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
    const sourceFile = ts.createSourceFile(filePath, sourceText, ts.ScriptTarget.Latest, true, scriptKind);
    const relativePath = path.relative(frontendDirectory, filePath).replaceAll('\\', '/');
    const findings = [];

    const addFinding = (kind, value) => {
        const text = normalize(value);

        if (text && containsEnglish(text) && !/^[a-z][\w-]*(?:\.[\w-]+)+$/.test(text)) {
            findings.push(`${relativePath}\t${kind}\t${text}`);
        }
    };

    const addChineseDictionaryFinding = (node) => {
        for (const value of expressionStrings(node)) {
            const text = normalize(value);
            const textWithoutPlaceholders = text.replaceAll(/\{\w+\}/g, '');

            if (containsEnglish(textWithoutPlaceholders) && !containsChinese(textWithoutPlaceholders)) {
                findings.push(`${relativePath}\tdictionary-value\t${text}`);
            }
        }
    };

    const visit = (node) => {
        if (ts.isJsxText(node)) {
            addFinding('jsx-text', node.text);
        } else if (ts.isJsxAttribute(node) && visibleAttributes.has(node.name.text)) {
            if (node.initializer && ts.isStringLiteral(node.initializer)) {
                addFinding(`attribute:${node.name.text}`, node.initializer.text);
            } else if (node.initializer && ts.isJsxExpression(node.initializer) && node.initializer.expression) {
                for (const value of expressionStrings(node.initializer.expression)) {
                    addFinding(`attribute:${node.name.text}`, value);
                }
            }
        } else if (ts.isJsxExpression(node) && node.expression && !ts.isJsxAttribute(node.parent)) {
            for (const value of expressionStrings(node.expression)) {
                addFinding('jsx-expression', value);
            }
        } else if (ts.isPropertyAssignment(node)) {
            if (relativePath === 'src/lib/i18n/locales/zh-CN.ts') {
                addChineseDictionaryFinding(node.initializer);
            } else if (visibleProperties.has(propertyName(node.name))) {
                for (const value of expressionStrings(node.initializer)) {
                    addFinding(`property:${propertyName(node.name)}`, value);
                }
            }
        } else if (ts.isVariableDeclaration(node) && node.initializer) {
            if (ts.isIdentifier(node.name) && visibleVariableName(node.name.text)) {
                for (const value of expressionStrings(node.initializer)) {
                    addFinding(`variable:${node.name.text}`, value);
                }
            } else if (
                ts.isArrayBindingPattern(node.name) &&
                ts.isCallExpression(node.initializer) &&
                node.initializer.expression.getText(sourceFile) === 'useState' &&
                node.initializer.arguments[0]
            ) {
                const stateName = node.name.elements[0];

                if (
                    ts.isBindingElement(stateName) &&
                    ts.isIdentifier(stateName.name) &&
                    visibleVariableName(stateName.name.text)
                ) {
                    for (const value of expressionStrings(node.initializer.arguments[0])) {
                        addFinding(`state:${stateName.name.text}`, value);
                    }
                }
            }
        } else if (ts.isCallExpression(node)) {
            const callee = node.expression.getText(sourceFile);
            const isVisibleNotification =
                callee === 'toast' ||
                ['toast.error', 'toast.info', 'toast.success', 'toast.warning'].includes(callee) ||
                ['alert', 'confirm', 'window.alert', 'window.confirm'].includes(callee);
            const isVisibleSetter = /^(set|update)[A-Z].*(Description|Error|Label|Message|Text|Title|Tooltip)$/.test(
                callee,
            );

            if ((isVisibleNotification || isVisibleSetter) && node.arguments[0]) {
                for (const value of expressionStrings(node.arguments[0])) {
                    addFinding(`call:${callee}`, value);
                }
            }
        }

        ts.forEachChild(node, visit);
    };

    visit(sourceFile);

    return findings;
};

const files = await collectFiles(sourceDirectory);
const currentFindings = [...new Set((await Promise.all(files.map(scanFile))).flat())].sort();

const projectSurfaceErrors = [];
const indexHtml = await readFile(path.join(frontendDirectory, 'index.html'), 'utf8');

if (!/<html\s+lang=["']zh-CN["']/.test(indexHtml)) {
    projectSurfaceErrors.push('frontend/index.html 必须声明 lang="zh-CN"。');
}

const dashboardPaths = [
    'observability/grafana/dashboards/home.json',
    'observability/grafana/dashboards/components/pentagi_service.json',
];
const isTechnicalDashboardTitle = (value) => /^[a-z][a-z0-9_]*(?:\([^)]*\))?$/.test(value);

const visitDashboard = (value, relativePath) => {
    if (Array.isArray(value)) {
        value.forEach((item) => visitDashboard(item, relativePath));
        return;
    }

    if (!value || typeof value !== 'object') {
        return;
    }

    for (const [key, child] of Object.entries(value)) {
        if (
            (key === 'title' || key === 'description') &&
            typeof child === 'string' &&
            containsEnglish(child) &&
            !containsChinese(child) &&
            !isTechnicalDashboardTitle(child)
        ) {
            projectSurfaceErrors.push(`${relativePath} 的 ${key} 仍为英文：${normalize(child)}`);
        }

        visitDashboard(child, relativePath);
    }
};

for (const relativePath of dashboardPaths) {
    const content = await readFile(path.join(repositoryDirectory, relativePath), 'utf8');
    visitDashboard(JSON.parse(content), relativePath);
}

const datasourceConfig = await readFile(
    path.join(repositoryDirectory, 'observability/grafana/config/provisioning/datasources/datasource.yml'),
    'utf8',
);
const traceLabel = datasourceConfig.match(/^\s*urlDisplayLabel:\s*["']?([^\r\n"']+)/m)?.[1]?.trim();

if (traceLabel && containsEnglish(traceLabel) && !containsChinese(traceLabel)) {
    projectSurfaceErrors.push(`Grafana 日志追踪入口仍为英文：${traceLabel}`);
}

if (projectSurfaceErrors.length > 0) {
    console.error('\n项目自有中文入口检查失败：');
    projectSurfaceErrors.forEach((error) => console.error(`- ${error}`));
    process.exit(1);
}

if (updateBaseline) {
    const content = [
        '# Generated by `pnpm i18n:baseline`.',
        '# Each entry is: relative path<TAB>detection kind<TAB>English text.',
        '# Entries may intentionally preserve brands, technical literals, source keys, or detector false positives.',
        '# Reviewed categories: product/provider names; CSS and code literals; paths and example values;',
        '# shared component English fallbacks overridden by current localized callers; English preset lookup keys.',
        '# Remove entries when localized; keep or add entries only after reviewing them.',
        ...currentFindings,
        '',
    ].join('\n');

    await writeFile(baselinePath, content, 'utf8');
    console.log(`Updated UI English baseline with ${currentFindings.length} findings.`);
    process.exit(0);
}

const baselineText = await readFile(baselinePath, 'utf8');
const baselineFindings = baselineText
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line && !line.startsWith('#'))
    .sort();
const baselineSet = new Set(baselineFindings);
const currentSet = new Set(currentFindings);
const added = currentFindings.filter((finding) => !baselineSet.has(finding));
const removed = baselineFindings.filter((finding) => !currentSet.has(finding));

if (added.length === 0 && removed.length === 0) {
    console.log(`UI English baseline is current (${currentFindings.length} known findings).`);
    process.exit(0);
}

if (added.length > 0) {
    console.error('\nNew user-visible English requires translation or an explicit baseline review:');
    added.forEach((finding) => console.error(`+ ${finding}`));
}

if (removed.length > 0) {
    console.error('\nBaseline entries no longer found; regenerate it after confirming the localization changes:');
    removed.forEach((finding) => console.error(`- ${finding}`));
}

console.error('\nRun `pnpm i18n:baseline` only after reviewing every reported change.');
process.exit(1);
