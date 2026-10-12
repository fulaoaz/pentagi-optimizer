export const formatPromptId = (key: string): string => {
    if (!/^[a-z][a-zA-Z]*$/.test(key)) {
        return t ? t('title.prompt') : 'Prompt';
    }

    if (t) {
        return translatePromptName(key, t);
    }

    return key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
};
