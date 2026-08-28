import type { Translate } from '@/lib/i18n';

export const providerTestNameKeys: Record<string, string> = {
    'Ask Advice Function': 'settings.provider.testName.askAdviceFunction',
    'Basic Context Memory Test': 'settings.provider.testName.basicContextMemory',
    'Basic Echo Function': 'settings.provider.testName.basicEchoFunction',
    'Basic Echo Function Streaming': 'settings.provider.testName.basicEchoFunctionStreaming',
    'Count from 1 to 3 Streaming': 'settings.provider.testName.countOneToThreeStreaming',
    'Count from 1 to 5': 'settings.provider.testName.countOneToFive',
    'Cybersecurity Workflow Memory Test': 'settings.provider.testName.cybersecurityWorkflowMemory',
    'Function Argument Memory Test': 'settings.provider.testName.functionArgumentMemory',
    'Function Response Memory Test': 'settings.provider.testName.functionResponseMemory',
    'JSON Response Function': 'settings.provider.testName.jsonResponseFunction',
    'Math Calculation': 'settings.provider.testName.mathCalculation',
    'Penetration Testing Framework': 'settings.provider.testName.pentestFramework',
    'Penetration Testing Memory with Tool Call': 'settings.provider.testName.pentestMemoryWithToolCall',
    'Penetration Testing Methodology': 'settings.provider.testName.pentestMethodology',
    'Penetration Testing Tool Selection': 'settings.provider.testName.pentestToolSelection',
    'Person Information JSON': 'settings.provider.testName.personInformationJson',
    'Person Information JSON Streaming': 'settings.provider.testName.personInformationJsonStreaming',
    'Project Information JSON': 'settings.provider.testName.projectInformationJson',
    'Search Query Function': 'settings.provider.testName.searchQueryFunction',
    'Search Query Function Streaming': 'settings.provider.testName.searchQueryFunctionStreaming',
    'Simple Math': 'settings.provider.testName.simpleMath',
    'Simple Math Streaming': 'settings.provider.testName.simpleMathStreaming',
    'SQL Injection Attack Type': 'settings.provider.testName.sqlInjectionAttackType',
    'Text Transform Uppercase': 'settings.provider.testName.textTransformUppercase',
    'User Profile JSON': 'settings.provider.testName.userProfileJson',
    'Vulnerability Assessment Tools': 'settings.provider.testName.vulnerabilityAssessmentTools',
    'Vulnerability Report Memory Test': 'settings.provider.testName.vulnerabilityReportMemory',
    'Web Application Security Scanner': 'settings.provider.testName.webApplicationSecurityScanner',
};

const providerTestTypeKeys: Record<string, string> = {
    completion: 'settings.provider.testType.completion',
    tool: 'settings.provider.testType.tool',
};

export const translateProviderTestName = (name: string, t: Translate): string => {
    const executionErrorPrefix = 'Execution Error:';

    if (name.startsWith(executionErrorPrefix)) {
        return t('settings.provider.testName.executionError', {
            error: name.slice(executionErrorPrefix.length).trim(),
        });
    }

    const key = providerTestNameKeys[name];

    return key ? t(key) : name;
};

export const translateProviderTestType = (type: string, t: Translate): string => {
    if (type === 'json') {
        return 'JSON';
    }

    const key = providerTestTypeKeys[type];

    return key ? t(key) : type;
};
