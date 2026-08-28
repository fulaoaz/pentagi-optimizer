import { useContext } from 'react';

import { LocaleProviderContext } from '@/providers/locale-provider';

/**
 * Access the active locale and the `t()` translator.
 *
 * ```tsx
 * const { t } = useLocale();
 * <Button>{t('common.save')}</Button>
 * ```
 */
export const useLocale = () => useContext(LocaleProviderContext);
