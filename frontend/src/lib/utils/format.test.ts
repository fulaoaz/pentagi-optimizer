import { describe, expect, it } from 'vitest';

import { formatDuration } from './format';

describe('formatDuration', () => {
    it('formats durations with Chinese units', () => {
        expect(formatDuration(3661, 'zh-CN')).toBe('1 小时 1 分钟');
        expect(formatDuration(61, 'zh-CN')).toBe('1 分钟 1 秒');
        expect(formatDuration(1.25, 'zh-CN')).toBe('1.3 秒');
        expect(formatDuration(0.25, 'zh-CN')).toBe('250 毫秒');
    });

    it('keeps the existing English format by default', () => {
        expect(formatDuration(3661)).toBe('1h 1m');
    });
});
