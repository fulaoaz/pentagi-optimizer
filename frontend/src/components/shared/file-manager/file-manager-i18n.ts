import type { Locale, Translate } from '@/lib/i18n';

import type { FileManagerLabels, FileManagerSortColumn } from './file-manager-types';

import { formatModifiedRelative } from './file-manager-utils';

const columnKey: Record<FileManagerSortColumn, string> = {
    modified: 'fileManager.columnModified',
    name: 'fileManager.columnName',
    size: 'fileManager.columnSize',
};

export const buildFileManagerLabels = (locale: Locale, t: Translate): FileManagerLabels => ({
    bulkCancel: t('fileManager.bulkCancel'),
    bulkMoreActions: t('fileManager.bulkMoreActions'),
    collapseAllAriaLabel: t('fileManager.collapseAll'),
    columnModified: t('fileManager.columnModified'),
    columnName: t('fileManager.columnName'),
    columnSize: t('fileManager.columnSize'),
    expandAllAriaLabel: t('fileManager.expandAll'),
    formatModified: (modifiedAt) => formatModifiedRelative(modifiedAt, locale),
    pluralizeItems: (count) => t(count === 1 ? 'fileManager.itemCountOne' : 'fileManager.itemCountMany', { count }),
    selectAllAriaLabel: t('fileManager.selectAll'),
    selectedLabel: (count) => t('fileManager.selectedCount', { count }),
    sortHeaderAriaLabel: (column, direction) => {
        const values = { label: t(columnKey[column]) };

        if (direction === 'asc') {
            return t('fileManager.sortDescending', values);
        }

        if (direction === 'desc') {
            return t('fileManager.clearSorting', values);
        }

        return t('fileManager.sortAscending', values);
    },
});
