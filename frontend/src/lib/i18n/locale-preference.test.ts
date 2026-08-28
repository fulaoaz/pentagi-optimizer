import { describe, expect, it } from 'vitest';

import { resolveLocalePreference } from './locale-preference';

describe('resolveLocalePreference', () => {
    it('uses a stored locale before browser preferences', () => {
        expect(resolveLocalePreference('en', ['zh-CN'])).toBe('en');
    });

    it('defaults to Simplified Chinese for a new visitor', () => {
        expect(resolveLocalePreference(null, ['zh-Hans-SG', 'en-US'])).toBe('zh-CN');
    });

    it('does not let browser preferences override the Chinese project default', () => {
        expect(resolveLocalePreference(undefined, ['en-GB', 'zh-CN'])).toBe('zh-CN');
    });
});
