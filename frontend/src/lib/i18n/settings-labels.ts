import type { Translate } from './types';

const agentNameKeys: Record<string, string> = {
    adviser: 'agent.adviser',
    assistant: 'agent.assistant',
    coder: 'agent.coder',
    enricher: 'agent.enricher',
    generator: 'agent.generator',
    installer: 'agent.installer',
    memorist: 'agent.memorist',
    pentester: 'agent.pentester',
    primaryAgent: 'agent.primaryAgent',
    refiner: 'agent.refiner',
    reflector: 'agent.reflector',
    reporter: 'agent.reporter',
    searcher: 'agent.searcher',
    simple: 'agent.simple',
    simpleJson: 'agent.simpleJson',
    summarizer: 'agent.summarizer',
    toolCallFixer: 'agent.toolCallFixer',
};

const promptNameKeys: Record<string, string> = {
    chooseDockerImage: 'promptName.chooseDockerImage',
    chooseUserLanguage: 'promptName.chooseUserLanguage',
    collectToolCallID: 'promptName.collectToolCallID',
    detectToolCallIDPattern: 'promptName.detectToolCallIDPattern',
    getExecutionLogs: 'promptName.getExecutionLogs',
    getFlowDescription: 'promptName.getFlowDescription',
    getFullExecutionContext: 'promptName.getFullExecutionContext',
    getShortExecutionContext: 'promptName.getShortExecutionContext',
    getTaskDescription: 'promptName.getTaskDescription',
    questionExecutionMonitor: 'promptName.questionExecutionMonitor',
    questionTaskPlanner: 'promptName.questionTaskPlanner',
    taskAssignmentWrapper: 'promptName.taskAssignmentWrapper',
};

const providerFieldKeys: Record<string, string> = {
    frequencyPenalty: 'settings.provider.frequencyPenalty',
    maxLength: 'settings.provider.maxLength',
    maxTokens: 'settings.provider.maxTokens',
    minLength: 'settings.provider.minLength',
    model: 'settings.provider.model',
    name: 'settings.provider.name',
    presencePenalty: 'settings.provider.presencePenalty',
    'price.cacheRead': 'settings.provider.cacheReadPrice',
    'price.cacheWrite': 'settings.provider.cacheWritePrice',
    'price.input': 'settings.provider.inputPrice',
    'price.output': 'settings.provider.outputPrice',
    'reasoning.effort': 'settings.provider.reasoningEffort',
    'reasoning.maxTokens': 'settings.provider.reasoningMaxTokens',
    repetitionPenalty: 'settings.provider.repetitionPenalty',
    temperature: 'settings.provider.temperature',
    topK: 'settings.provider.topK',
    topP: 'settings.provider.topP',
    type: 'settings.provider.type',
};

const formatCamelCase = (value: string): string =>
    value.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (character) => character.toUpperCase());

export const translateAgentName = (name: string, t: Translate): string => {
    const normalizedName = name.replaceAll(/_([a-z])/g, (_match, character: string) => character.toUpperCase());
    const key = agentNameKeys[name] ?? agentNameKeys[normalizedName];

    return key ? t(key) : formatCamelCase(normalizedName);
};

export const translatePromptName = (name: string, t: Translate): string => {
    const key = agentNameKeys[name] ?? promptNameKeys[name];

    return key ? t(key) : formatCamelCase(name);
};

export const translateProviderFieldName = (path: string, t: Translate): string => {
    const key = providerFieldKeys[path];

    return key ? t(key) : formatCamelCase(path.split('.').at(-1) ?? path);
};

export const translateProviderFieldPath = (fieldPath: string, t: Translate): string => {
    const parts = fieldPath.split('.');

    if (parts[0] !== 'agents') {
        return translateProviderFieldName(fieldPath, t);
    }

    const labels = [t('settings.provider.agentConfigurations')];
    const agentName = parts[1];

    if (agentName) {
        labels.push(translateAgentName(agentName, t));
    }

    const configPath = parts.slice(2).join('.');

    if (configPath) {
        labels.push(translateProviderFieldName(configPath, t));
    }

    return labels.join(' → ');
};
