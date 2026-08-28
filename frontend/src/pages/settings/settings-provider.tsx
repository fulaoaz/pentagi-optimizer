import { zodResolver } from '@hookform/resolvers/zod';
import {
    AlertCircle,
    Check,
    CheckCircle,
    ChevronsUpDown,
    Clock,
    Cpu,
    Lightbulb,
    Loader2,
    Play,
    Save,
    Trash2,
    XCircle,
} from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useController, useForm, useFormState, useWatch } from 'react-hook-form';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { z } from 'zod';

import type {
    AgentConfigInput,
    AgentsConfigInput,
    ProviderConfigFragmentFragment,
    ProviderType,
} from '@/graphql/types';
import type { Translate } from '@/lib/i18n';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { FormSubmitButton } from '@/components/ui/form-submit-button';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { StatusCard } from '@/components/ui/status-card';
import {
    AgentConfigType,
    ReasoningEffort,
    useCreateProviderMutation,
    useDeleteProviderMutation,
    useSettingsProvidersQuery,
    useTestAgentMutation,
    useTestProviderMutation,
    useUpdateProviderMutation,
} from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';
import { translateAgentName, translateProviderFieldPath } from '@/lib/i18n/settings-labels';
import { cn } from '@/lib/utils';

import { translateProviderTestName, translateProviderTestType } from './provider-test-labels';

interface BaseFieldProps extends ControllerProps {
    label: string;
}

interface BaseInputProps {
    placeholder?: string;
}

interface ControllerProps {
    control: any;
    disabled?: boolean;
    name: string;
}

interface FormComboboxItemProps extends BaseFieldProps, BaseInputProps {
    allowCustom?: boolean;
    contentClass?: string;
    description?: string;
    options: string[];
}

interface FormInputNumberItemProps extends BaseFieldProps, NumberInputProps {
    description?: string;
    valueType?: 'float' | 'integer';
}

interface FormInputStringItemProps extends BaseFieldProps, BaseInputProps {
    description?: string;
}

interface FormModelComboboxItemProps extends BaseFieldProps, BaseInputProps {
    allowCustom?: boolean;
    contentClass?: string;
    description?: string;
    onOptionSelect?: (option: ModelOption) => void;
    options: ModelOption[];
}

interface ModelOption {
    name: string;
    price?: null | { cacheRead: number; cacheWrite: number; input: number; output: number };
    thinking?: boolean;
}

interface NumberInputProps extends BaseInputProps {
    max?: string;
    min?: string;
    step?: string;
}

type Provider = ProviderConfigFragmentFragment;

function FormComboboxItem({
    allowCustom = true,
    contentClass,
    control,
    description,
    disabled,
    label,
    name,
    options,
    placeholder,
}: FormComboboxItemProps) {
    const { t } = useLocale();
    const { field, fieldState } = useController({
        control,
        defaultValue: undefined,
        disabled,
        name,
    });

    const [isOpen, setIsOpen] = useState(false);
    const [search, setSearch] = useState('');

    const filteredOptions = options.filter((option) => option?.toLowerCase().includes(search?.toLowerCase()));

    const displayValue = field.value ?? '';

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Popover
                    onOpenChange={setIsOpen}
                    open={isOpen}
                >
                    <PopoverTrigger asChild>
                        <Button
                            className={cn('w-full justify-between', !displayValue && 'text-muted-foreground')}
                            disabled={disabled}
                            variant="outline"
                        >
                            {displayValue || placeholder}
                            <ChevronsUpDown className="opacity-50" />
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent
                        align="start"
                        className={cn(contentClass, 'p-0')}
                        style={{
                            maxHeight: 'var(--radix-popover-content-available-height)',
                            width: 'var(--radix-popover-trigger-width)',
                        }}
                    >
                        <Command>
                            <CommandInput
                                className="h-9"
                                onValueChange={setSearch}
                                placeholder={t('settings.provider.searchField', { field: label.toLocaleLowerCase() })}
                                value={search}
                            />
                            <CommandList>
                                <CommandEmpty>
                                    <div className="py-2 text-center">
                                        <p className="text-muted-foreground text-sm">
                                            {t('settings.provider.noFieldFound', {
                                                field: label.toLocaleLowerCase(),
                                            })}
                                        </p>
                                        {search && allowCustom && (
                                            <Button
                                                className="mt-2"
                                                onClick={() => {
                                                    field.onChange(search);
                                                    setIsOpen(false);
                                                    setSearch('');
                                                }}
                                                size="sm"
                                                variant="ghost"
                                            >
                                                {t('settings.provider.useCustomValue', {
                                                    field: label.toLocaleLowerCase(),
                                                    value: search,
                                                })}
                                            </Button>
                                        )}
                                    </div>
                                </CommandEmpty>
                                <CommandGroup>
                                    {filteredOptions.map((option) => (
                                        <CommandItem
                                            key={option}
                                            onSelect={() => {
                                                field.onChange(option);
                                                setIsOpen(false);
                                                setSearch('');
                                            }}
                                            value={option}
                                        >
                                            {option}
                                            <Check
                                                className={cn(
                                                    'ml-auto',
                                                    displayValue === option ? 'opacity-100' : 'opacity-0',
                                                )}
                                            />
                                        </CommandItem>
                                    ))}
                                </CommandGroup>
                            </CommandList>
                        </Command>
                    </PopoverContent>
                </Popover>
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
}

function FormInputNumberItem({
    control,
    description,
    disabled,
    label,
    max,
    min,
    name,
    placeholder,
    step,
    valueType = 'float',
}: FormInputNumberItemProps) {
    const { field, fieldState } = useController({
        control,
        defaultValue: undefined,
        disabled,
        name,
    });

    const parseValue = (value: string) => {
        if (value === '') {
            return null;
        }

        return valueType === 'float' ? Number.parseFloat(value) : Number.parseInt(value);
    };

    const inputProps = {
        max,
        min,
        placeholder,
        step,
        type: 'number' as const,
    };

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Input
                    {...field}
                    {...inputProps}
                    onChange={(event) => {
                        const { value } = event.target;
                        field.onChange(parseValue(value));
                    }}
                    value={field.value ?? ''}
                />
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
}

function FormInputStringItem({ control, description, disabled, label, name, placeholder }: FormInputStringItemProps) {
    const { field, fieldState } = useController({
        control,
        defaultValue: undefined,
        disabled,
        name,
    });

    const inputProps = { placeholder };

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Input
                    {...field}
                    {...inputProps}
                    value={field.value ?? ''}
                />
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
}

function FormModelComboboxItem({
    allowCustom = true,
    contentClass,
    control,
    description,
    disabled,
    label,
    name,
    onOptionSelect,
    options,
    placeholder,
}: FormModelComboboxItemProps) {
    const { t } = useLocale();
    const { field, fieldState } = useController({
        control,
        defaultValue: undefined,
        disabled,
        name,
    });

    const [isOpen, setIsOpen] = useState(false);
    const [search, setSearch] = useState('');

    const filteredOptions = options.filter((option) => option.name?.toLowerCase().includes(search?.toLowerCase()));

    const displayValue = field.value ?? '';

    const formatPrice = (
        price?: null | { cacheRead: number; cacheWrite: number; input: number; output: number },
    ): string => {
        if (!price || ((!price.input || price.input === 0) && (!price.output || price.output === 0))) {
            return t('settings.provider.free');
        }

        const formatValue = (value: number): string => {
            return value.toFixed(6).replace(/\.?0+$/, '');
        };

        const basePrice = `$${formatValue(price.input)}/$${formatValue(price.output)}`;

        const hasCachePrices = (price.cacheRead && price.cacheRead > 0) || (price.cacheWrite && price.cacheWrite > 0);

        if (hasCachePrices) {
            const cacheParts: string[] = [];

            if (price.cacheRead && price.cacheRead > 0) {
                cacheParts.push(`R:$${formatValue(price.cacheRead)}`);
            }

            if (price.cacheWrite && price.cacheWrite > 0) {
                cacheParts.push(`W:$${formatValue(price.cacheWrite)}`);
            }

            return `${basePrice} (${cacheParts.join(', ')})`;
        }

        return basePrice;
    };

    return (
        <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
                <Popover
                    onOpenChange={setIsOpen}
                    open={isOpen}
                >
                    <div className="flex w-full">
                        {/* Input field - main control */}
                        <Input
                            className="rounded-r-none border-r-0 focus-visible:z-10"
                            disabled={disabled}
                            onChange={(event) => field.onChange(event.target.value)}
                            placeholder={placeholder}
                            value={displayValue}
                        />
                        {/* Dropdown trigger button */}
                        <PopoverTrigger asChild>
                            <Button
                                className="rounded-l-none border-l-0 px-3 hover:z-10"
                                disabled={disabled}
                                type="button"
                                variant="outline"
                            >
                                <ChevronsUpDown className="size-4 opacity-50" />
                            </Button>
                        </PopoverTrigger>
                        <PopoverContent
                            align="end"
                            className={cn(contentClass, 'w-80 p-0 sm:w-[480px] md:w-[640px]')}
                        >
                            <Command>
                                <CommandInput
                                    className="h-9"
                                    onValueChange={setSearch}
                                    placeholder={t('settings.provider.searchField', {
                                        field: label.toLocaleLowerCase(),
                                    })}
                                    value={search}
                                />
                                <CommandList>
                                    <CommandEmpty>
                                        <div className="py-2 text-center">
                                            <p className="text-muted-foreground text-sm">
                                                {t('settings.provider.noFieldFound', {
                                                    field: label.toLocaleLowerCase(),
                                                })}
                                            </p>
                                            {search && allowCustom && (
                                                <Button
                                                    className="mt-2"
                                                    onClick={() => {
                                                        field.onChange(search);
                                                        setIsOpen(false);
                                                        setSearch('');
                                                    }}
                                                    size="sm"
                                                    variant="ghost"
                                                >
                                                    {t('settings.provider.useCustomValue', {
                                                        field: label.toLocaleLowerCase(),
                                                        value: search,
                                                    })}
                                                </Button>
                                            )}
                                        </div>
                                    </CommandEmpty>
                                    <CommandGroup>
                                        {filteredOptions.map((option) => (
                                            <CommandItem
                                                key={option.name}
                                                onSelect={() => {
                                                    field.onChange(option.name);
                                                    onOptionSelect?.(option);
                                                    setIsOpen(false);
                                                    setSearch('');
                                                }}
                                                value={option.name}
                                            >
                                                <div className="flex w-full min-w-0 items-center justify-between gap-2">
                                                    <div className="flex min-w-0 items-center gap-2">
                                                        <span className="truncate">{option.name}</span>
                                                        {option.thinking && (
                                                            <Lightbulb className="text-muted-foreground size-3" />
                                                        )}
                                                    </div>
                                                    <span className="text-muted-foreground shrink-0 text-xs whitespace-nowrap">
                                                        {formatPrice(option.price)}
                                                    </span>
                                                </div>
                                                <Check
                                                    className={cn(
                                                        'ml-auto',
                                                        displayValue === option.name ? 'opacity-100' : 'opacity-0',
                                                    )}
                                                />
                                            </CommandItem>
                                        ))}
                                    </CommandGroup>
                                </CommandList>
                            </Command>
                        </PopoverContent>
                    </div>
                </Popover>
            </FormControl>
            {description && <FormDescription>{description}</FormDescription>}
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
}

const buildAgentConfigSchema = (t: Translate) =>
    z
        .object({
            frequencyPenalty: z.preprocess(
                (value) => (value === '' || value === undefined ? null : value),
                z.number().nullable().optional(),
            ),
            maxLength: z.preprocess(
                (value) => (value === '' || value === undefined ? null : value),
                z.number().nullable().optional(),
            ),
            maxTokens: z.preprocess(
                (value) => (value === '' || value === undefined ? null : value),
                z.number().nullable().optional(),
            ),
            minLength: z.preprocess(
                (value) => (value === '' || value === undefined ? null : value),
                z.number().nullable().optional(),
            ),
            model: z.preprocess((value) => value || '', z.string().min(1, t('settings.provider.modelRequired'))),
            presencePenalty: z.preprocess(
                (value) => (value === '' || value === undefined ? null : value),
                z.number().nullable().optional(),
            ),
            price: z
                .object({
                    cacheRead: z.preprocess(
                        (value) => (value === '' || value === undefined ? null : value),
                        z.number().nullable().optional(),
                    ),
                    cacheWrite: z.preprocess(
                        (value) => (value === '' || value === undefined ? null : value),
                        z.number().nullable().optional(),
                    ),
                    input: z.preprocess(
                        (value) => (value === '' || value === undefined ? null : value),
                        z.number().nullable().optional(),
                    ),
                    output: z.preprocess(
                        (value) => (value === '' || value === undefined ? null : value),
                        z.number().nullable().optional(),
                    ),
                })
                .nullable()
                .optional(),
            reasoning: z
                .object({
                    effort: z.preprocess(
                        (value) => (value === '' || value === undefined ? null : value),
                        z.string().nullable().optional(),
                    ),
                    maxTokens: z.preprocess(
                        (value) => (value === '' || value === undefined ? null : value),
                        z.number().nullable().optional(),
                    ),
                })
                .nullable()
                .optional(),
            repetitionPenalty: z.preprocess(
                (value) => (value === '' || value === undefined ? null : value),
                z.number().nullable().optional(),
            ),
            temperature: z.preprocess(
                (value) => (value === '' || value === undefined ? null : value),
                z.number().nullable().optional(),
            ),
            topK: z.preprocess(
                (value) => (value === '' || value === undefined ? null : value),
                z.number().nullable().optional(),
            ),
            topP: z.preprocess(
                (value) => (value === '' || value === undefined ? null : value),
                z.number().nullable().optional(),
            ),
        })
        .optional();

const buildFormSchema = (t: Translate) =>
    z.object({
        agents: z.record(z.string(), buildAgentConfigSchema(t)).optional(),
        name: z.preprocess(
            (value) => value || '',
            z.string().min(1, t('settings.provider.nameRequired')).max(50, t('settings.provider.maxCharacters')),
        ),
        type: z.preprocess((value) => value || '', z.string().min(1, t('settings.provider.typeRequired'))),
    });

type FormAgents = FormData['agents'];

type FormData = z.infer<ReturnType<typeof buildFormSchema>>;

const getReasoningEffort = (effort: null | string | undefined): null | ReasoningEffort => {
    if (!effort) {
        return null;
    }

    switch (effort.toLowerCase()) {
        case 'high': {
            return ReasoningEffort.High;
        }

        case 'low': {
            return ReasoningEffort.Low;
        }

        case 'medium': {
            return ReasoningEffort.Medium;
        }

        default: {
            return null;
        }
    }
};

const transformFormToGraphQL = (
    formData: FormData,
): {
    agents: AgentsConfigInput;
    name: string;
    type: ProviderType;
} => {
    const agents = Object.entries(formData.agents || {})
        .filter(([key, data]) => key !== '__typename' && data?.model)
        .reduce((configs, [key, data]) => {
            const config: AgentConfigInput = {
                frequencyPenalty: data?.frequencyPenalty ?? null,
                maxLength: data?.maxLength ?? null,
                maxTokens: data?.maxTokens ?? null,
                minLength: data?.minLength ?? null,
                model: data!.model,
                presencePenalty: data?.presencePenalty ?? null,
                price:
                    data?.price &&
                    typeof data?.price.input === 'number' &&
                    typeof data?.price.output === 'number' &&
                    typeof data?.price.cacheRead === 'number' &&
                    typeof data?.price.cacheWrite === 'number'
                        ? {
                              cacheRead: data.price.cacheRead,
                              cacheWrite: data.price.cacheWrite,
                              input: data.price.input,
                              output: data.price.output,
                          }
                        : null,
                reasoning: data?.reasoning
                    ? {
                          effort: getReasoningEffort(data?.reasoning.effort),
                          maxTokens: data?.reasoning.maxTokens ?? null,
                      }
                    : null,
                repetitionPenalty: data?.repetitionPenalty ?? null,
                temperature: data?.temperature ?? null,
                topK: data?.topK ?? null,
                topP: data?.topP ?? null,
            };

            return { ...configs, [key]: config };
        }, {} as AgentsConfigInput);

    return {
        agents,
        name: formData.name,
        type: formData.type as ProviderType,
    };
};

const normalizeGraphQLData = (obj: unknown): unknown => {
    if (obj === null || obj === undefined) {
        return obj;
    }

    if (Array.isArray(obj)) {
        return obj.map(normalizeGraphQLData);
    }

    if (typeof obj === 'object') {
        return Object.fromEntries(
            Object.entries(obj)
                .filter(([key]) => key !== '__typename')
                .map(([key, value]) => [key, normalizeGraphQLData(value)]),
        );
    }

    return obj;
};

interface TestResultsDialogProps {
    handleOpenChange: (isOpen: boolean) => void;
    isOpen: boolean;
    results: any;
}

function TestResultsDialog({ handleOpenChange, isOpen, results }: TestResultsDialogProps) {
    const { t } = useLocale();

    if (!results) {
        return null;
    }

    const agentResults = Object.entries(results)
        .filter(([key]) => key !== '__typename')
        .map(([agentType, agentData]: [string, any]) => ({
            agentType,
            tests: agentData?.tests || [],
        }));

    const getStatusIcon = (result: boolean) => {
        if (result === true) {
            return <CheckCircle className="size-4 text-green-500" />;
        } else if (result === false) {
            return <XCircle className="size-4 text-red-500" />;
        } else {
            return <Clock className="size-4 text-yellow-500" />;
        }
    };

    const getStatusColor = (result: boolean) => {
        if (result === true) {
            return 'text-green-600';
        } else if (result === false) {
            return 'text-red-600';
        } else {
            return 'text-yellow-600';
        }
    };

    return (
        <Dialog
            onOpenChange={handleOpenChange}
            open={isOpen}
        >
            <DialogContent className="flex max-h-[80vh] max-w-4xl flex-col">
                <DialogHeader className="shrink-0">
                    <DialogTitle>{t('settings.provider.testResults')}</DialogTitle>
                </DialogHeader>
                <div className="flex flex-1 flex-col gap-6 overflow-y-auto">
                    <Accordion
                        className="w-full"
                        type="multiple"
                    >
                        {agentResults.map(({ agentType, tests }) => {
                            const testsCount = tests.length;
                            const successTestsCount = tests.filter((test: any) => test.result === true).length;

                            return (
                                <AccordionItem
                                    key={agentType}
                                    value={agentType}
                                >
                                    <AccordionTrigger className="text-left">
                                        <div className="mr-4 flex w-full items-center justify-between">
                                            <span className="text-lg font-semibold">
                                                {translateAgentName(agentType, t)}
                                            </span>
                                            <span className="text-muted-foreground text-sm">
                                                {t('settings.provider.testsPassed', {
                                                    passed: successTestsCount,
                                                    total: testsCount,
                                                })}
                                            </span>
                                        </div>
                                    </AccordionTrigger>
                                    <AccordionContent>
                                        <div className="flex flex-col gap-3 pt-2">
                                            {tests.map((test: any, index: number) => (
                                                <div
                                                    className="rounded-lg border p-3"
                                                    key={index}
                                                >
                                                    <div className="mb-2 flex items-start justify-between">
                                                        <div className="flex items-center gap-2">
                                                            {getStatusIcon(test.result)}
                                                            <span className="font-medium">
                                                                {translateProviderTestName(test.name, t)}
                                                            </span>
                                                            {test.type && (
                                                                <span className="text-muted-foreground text-sm">
                                                                    ({translateProviderTestType(test.type, t)})
                                                                </span>
                                                            )}
                                                        </div>
                                                        <div className="text-muted-foreground flex items-center gap-3 text-sm">
                                                            {test.reasoning !== undefined && (
                                                                <span>
                                                                    {t('settings.provider.reasoning')}{' '}
                                                                    {test.reasoning ? t('common.yes') : t('common.no')}
                                                                </span>
                                                            )}
                                                            {test.streaming !== undefined && (
                                                                <span>
                                                                    {t('settings.provider.streaming')}{' '}
                                                                    {test.streaming ? t('common.yes') : t('common.no')}
                                                                </span>
                                                            )}
                                                            {test.latency && (
                                                                <span>
                                                                    {t('settings.provider.latency')}{' '}
                                                                    {t('settings.provider.latencyValue', {
                                                                        value: test.latency,
                                                                    })}
                                                                </span>
                                                            )}
                                                        </div>
                                                    </div>
                                                    <div
                                                        className={`text-sm font-medium ${getStatusColor(test.result)}`}
                                                    >
                                                        {t('settings.provider.result')}{' '}
                                                        {test.result === true
                                                            ? t('common.success')
                                                            : test.result === false
                                                              ? t('common.failed')
                                                              : t('common.unknown')}
                                                    </div>
                                                    {test.error && (
                                                        <div className="mt-2 rounded border border-red-200 bg-red-50 p-2 text-sm text-red-700">
                                                            <strong>{t('common.error')}:</strong> {test.error}
                                                        </div>
                                                    )}
                                                </div>
                                            ))}
                                            {tests.length === 0 && (
                                                <div className="text-muted-foreground py-4 text-center">
                                                    {t('settings.provider.noAgentTests')}
                                                </div>
                                            )}
                                        </div>
                                    </AccordionContent>
                                </AccordionItem>
                            );
                        })}
                    </Accordion>
                </div>
            </DialogContent>
        </Dialog>
    );
}

const agentTypesMap: Record<string, AgentConfigType> = {
    adviser: AgentConfigType.Adviser,
    assistant: AgentConfigType.Assistant,
    coder: AgentConfigType.Coder,
    enricher: AgentConfigType.Enricher,
    generator: AgentConfigType.Generator,
    installer: AgentConfigType.Installer,
    pentester: AgentConfigType.Pentester,
    primaryAgent: AgentConfigType.PrimaryAgent,
    refiner: AgentConfigType.Refiner,
    reflector: AgentConfigType.Reflector,
    searcher: AgentConfigType.Searcher,
    simple: AgentConfigType.Simple,
    simpleJson: AgentConfigType.SimpleJson,
};

const extractAgentTypes = (agents: unknown): null | string[] => {
    if (!agents || typeof agents !== 'object') {
        return null;
    }

    const types = Object.entries(agents)
        .filter(([key, data]) => key !== '__typename' && data)
        .map(([key]) => key)
        .sort();

    return types.length > 0 ? types : null;
};

const formatValidationErrors = (errors: Record<string, unknown>, t: Translate): string => {
    const collectMessages = (value: unknown, path: string[]): string[] => {
        if (!value || typeof value !== 'object') {
            return [];
        }

        const error = value as Record<string, unknown>;

        if (typeof error.message === 'string') {
            return [`• ${translateProviderFieldPath(path.join('.'), t)}: ${error.message}`];
        }

        return Object.entries(error)
            .filter(([key]) => !['ref', 'type', 'types'].includes(key))
            .flatMap(([key, nestedError]) => collectMessages(nestedError, [...path, key]));
    };

    return Object.entries(errors)
        .flatMap(([field, error]) => collectMessages(error, [field]))
        .join('\n');
};

function SettingsProvider() {
    const { t } = useLocale();
    const { providerId } = useParams<{ providerId: string }>();
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const { data, error, loading } = useSettingsProvidersQuery();
    const [createProvider, { error: createError, loading: isCreateLoading }] = useCreateProviderMutation();
    const [updateProvider, { error: updateError, loading: isUpdateLoading }] = useUpdateProviderMutation();
    const [deleteProvider, { error: deleteError, loading: isDeleteLoading }] = useDeleteProviderMutation();
    const [testProvider, { error: testError, loading: isTestLoading }] = useTestProviderMutation();
    const [testAgent, { error: agentTestError, loading: isAgentTestLoading }] = useTestAgentMutation();
    const [currentAgentKey, setCurrentAgentKey] = useState<null | string>(null);
    const [submitError, setSubmitError] = useState<null | string>(null);
    const [isTestDialogOpen, setIsTestDialogOpen] = useState(false);
    const [testResults, setTestResults] = useState<any>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [isLeaveDialogOpen, setIsLeaveDialogOpen] = useState(false);
    const [pendingBrowserBack, setPendingBrowserBack] = useState(false);
    const allowBrowserLeaveRef = useRef(false);
    const hasPushedBlockerStateRef = useRef(false);

    const isNew = providerId === 'new';
    const isLoading = isCreateLoading || isUpdateLoading || isDeleteLoading;
    const formSchema = useMemo(() => buildFormSchema(t), [t]);

    const form = useForm<FormData>({
        defaultValues: {
            agents: {},
            name: undefined,
            type: undefined,
        },
        resolver: zodResolver(formSchema),
    });

    const { control, formState, handleSubmit: handleFormSubmit, reset, setValue, trigger, watch } = form;

    const { isDirty } = useFormState({ control });

    // Push a synthetic history entry while the form is dirty so a browser-back can be intercepted
    // by popstate below — react-router's blocker doesn't cover the native back gesture.
    useEffect(() => {
        if (isDirty && !hasPushedBlockerStateRef.current) {
            window.history.pushState({ __pentagiBlock__: true }, '');
            hasPushedBlockerStateRef.current = true;
        }
    }, [isDirty]);

    useEffect(() => {
        const handlePopState = () => {
            if (!isDirty) {
                return;
            }

            if (allowBrowserLeaveRef.current) {
                allowBrowserLeaveRef.current = false;

                return;
            }

            setPendingBrowserBack(true);
            setIsLeaveDialogOpen(true);
            // Restore the blocker entry so the user stays on the page until they confirm.
            window.history.forward();
        };

        window.addEventListener('popstate', handlePopState, { capture: true });

        return () => {
            window.removeEventListener('popstate', handlePopState, { capture: true });
        };
    }, [isDirty]);

    const selectedType = useWatch({ control, name: 'type' });

    const providerName = useWatch({ control, name: 'name' });

    const formQueryParams = useMemo(
        () => ({
            id: searchParams.get('id'),
            type: searchParams.get('type'),
        }),
        [searchParams],
    );

    const getAgentTypes = () => {
        const agentsSource =
            (isNew &&
                selectedType &&
                data?.settingsProviders?.default?.[selectedType as keyof typeof data.settingsProviders.default]
                    ?.agents) ||
            (!isNew &&
                providerId &&
                data?.settingsProviders?.userDefined?.find((p: Provider) => p.id == providerId)?.agents) ||
            (data?.settingsProviders?.default &&
                Object.values(data.settingsProviders.default).find((provider) => provider?.agents)?.agents) ||
            null;

        return extractAgentTypes(agentsSource) ?? Object.keys(agentTypesMap);
    };

    const agentTypes = getAgentTypes();

    const availableModels = useMemo(() => {
        if (!data?.settingsProviders?.models || !selectedType) {
            return [];
        }

        const { models } = data.settingsProviders;
        const providerModels = models[selectedType as keyof typeof models];

        if (!providerModels?.length) {
            return [];
        }

        return providerModels
            .map((model: any) => ({
                name: model.name,
                price: model.price
                    ? {
                          cacheRead: model.price.cacheRead ?? 0,
                          cacheWrite: model.price.cacheWrite ?? 0,
                          input: model.price.input ?? 0,
                          output: model.price.output ?? 0,
                      }
                    : null,
                thinking: model.thinking,
            }))
            .filter((model) => model.name)
            .sort((a, b) => a.name.localeCompare(b.name));
    }, [data, selectedType]);

    useEffect(() => {
        if (!isNew || !selectedType || !data?.settingsProviders?.default || availableModels.length === 0) {
            return;
        }

        const defaultProvider =
            data.settingsProviders.default[selectedType as keyof typeof data.settingsProviders.default];

        if (defaultProvider?.agents) {
            const agents = Object.fromEntries(
                Object.entries(defaultProvider.agents)
                    .filter(([key]) => key !== '__typename')
                    .map(([key, data]) => {
                        const agent = { ...data };

                        if (agent.model && !availableModels.find((m) => m.name === agent.model)) {
                            agent.model = availableModels[0]?.name || agent.model;
                        }

                        return [key, agent];
                    }),
            );

            setValue('agents', normalizeGraphQLData(agents) as FormAgents);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [availableModels, data, isNew, selectedType]);

    useEffect(() => {
        if (!isNew) {
            if (searchParams.size > 0) {
                setSearchParams({});
            }

            return;
        }

        const queryId = searchParams.get('id');

        if (queryId) {
            return;
        }

        const queryType = searchParams.get('type');

        if (!selectedType && queryType) {
            return;
        }

        setSearchParams((prev) => {
            const params = new URLSearchParams(prev);

            if (selectedType) {
                params.set('type', selectedType);
            } else {
                params.delete('type');
            }

            return params;
        });
    }, [selectedType, setSearchParams, isNew, searchParams]);

    useEffect(() => {
        if (!data?.settingsProviders) {
            return;
        }

        const providers = data.settingsProviders;

        if (isNew || !providerId) {
            const queryType = formQueryParams.type ?? undefined;
            const queryId = formQueryParams.id;

            if (queryId && data?.settingsProviders?.userDefined) {
                const sourceProvider = data.settingsProviders.userDefined.find((p: Provider) => p.id == queryId);

                if (sourceProvider) {
                    const { agents, name, type: sourceType } = sourceProvider;

                    reset({
                        agents: agents ? (normalizeGraphQLData(agents) as FormAgents) : {},
                        name: t('settings.provider.copyName', { name }),
                        type: sourceType ?? undefined,
                    });

                    return;
                }
            } else if (queryType && data?.settingsProviders?.default) {
                const defaultProvider =
                    data.settingsProviders.default[queryType as keyof typeof data.settingsProviders.default];

                reset({
                    agents: defaultProvider?.agents ? (normalizeGraphQLData(defaultProvider.agents) as FormAgents) : {},
                    name: undefined,
                    type: queryType,
                });
            }

            // Bail out of the empty-form reset when `selectedType` is set — the agent-filling
            // effect above is the source of truth in that case and would fight us.
            if (!selectedType) {
                reset({
                    agents: {},
                    name: undefined,
                    type: queryType,
                });
            }

            return;
        }

        const provider = providers.userDefined?.find((provider: Provider) => provider.id == providerId);

        if (!provider) {
            navigate('/settings/providers');

            return;
        }

        const { agents, name, type } = provider;

        reset({
            agents: agents ? (normalizeGraphQLData(agents) as FormAgents) : {},
            name: name || undefined,
            type: type || undefined,
        });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [data, formQueryParams, isNew, providerId, selectedType]);

    const handleSubmit = async () => {
        // watch() — not getValues() — because disabled fields must be included in the payload.
        const formData = watch();

        try {
            setSubmitError(null);

            const mutationData = transformFormToGraphQL(formData);

            if (isNew) {
                await createProvider({
                    refetchQueries: ['settingsProviders'],
                    variables: mutationData,
                });
            } else {
                await updateProvider({
                    refetchQueries: ['settingsProviders'],
                    variables: {
                        ...mutationData,
                        providerId: providerId!,
                    },
                });
            }

            navigate('/settings/providers');
        } catch (error) {
            console.error('Submit error:', error);
            setSubmitError(error instanceof Error ? error.message : t('settings.provider.saveError'));
        }
    };

    const handleDelete = () => {
        if (isNew || !providerId) {
            return;
        }

        setIsDeleteDialogOpen(true);
    };

    const handleConfirmDelete = async () => {
        if (isNew || !providerId) {
            return;
        }

        try {
            setSubmitError(null);

            await deleteProvider({
                refetchQueries: ['settingsProviders'],
                variables: { providerId },
            });

            navigate('/settings/providers');
        } catch (error) {
            console.error('Delete error:', error);
            setSubmitError(error instanceof Error ? error.message : t('settings.provider.deleteError'));
        }
    };

    const handleTest = async () => {
        const isValid = await trigger();

        if (!isValid) {
            const { errors } = formState;
            const errorMessages = formatValidationErrors(errors, t);

            setSubmitError(`${t('settings.provider.formValidationErrors')}\n\n${errorMessages}`);

            return;
        }

        try {
            setSubmitError(null);

            const formData = watch();
            const { agents, type } = transformFormToGraphQL(formData);
            const result = await testProvider({
                variables: {
                    agents,
                    type,
                },
            });

            setTestResults(result.data?.testProvider);
            setIsTestDialogOpen(true);
        } catch (error) {
            console.error('Test error:', error);
            setSubmitError(error instanceof Error ? error.message : t('settings.provider.testError'));
        }
    };

    const handleTestAgent = async (agentKey: string) => {
        const isValid = await trigger();

        if (!isValid) {
            const { errors } = formState;
            const errorMessages = formatValidationErrors(errors, t);

            setSubmitError(`${t('settings.provider.formValidationErrors')}\n\n${errorMessages}`);

            return;
        }

        try {
            setSubmitError(null);
            setCurrentAgentKey(agentKey);
            // watch() — not getValues() — because disabled fields must be included in the payload.
            const formData = watch();
            const { agents, type } = transformFormToGraphQL(formData);

            const agent = agents[agentKey as keyof AgentsConfigInput] as AgentConfigInput;

            const singleResult = await testAgent({
                variables: { agent, agentType: agentTypesMap[agentKey] ?? AgentConfigType.Simple, type },
            });
            setTestResults({ [agentKey]: singleResult.data?.testAgent });
            setIsTestDialogOpen(true);
            setCurrentAgentKey(null);

            return;
        } catch (error) {
            console.error('Test error:', error);
            setSubmitError(error instanceof Error ? error.message : t('settings.provider.testError'));
            setCurrentAgentKey(null);
        }
    };

    const handleBack = () => {
        if (isDirty) {
            setIsLeaveDialogOpen(true);

            return;
        }

        navigate('/settings/providers');
    };

    const handleConfirmLeave = () => {
        if (pendingBrowserBack) {
            allowBrowserLeaveRef.current = true;
            setPendingBrowserBack(false);
            // Step over the synthetic blocker entry into the actual previous page.
            window.history.go(-2);

            return;
        }

        navigate('/settings/providers');
    };

    const handleLeaveDialogOpenChange = (open: boolean) => {
        if (!open && pendingBrowserBack) {
            setPendingBrowserBack(false);
        }

        setIsLeaveDialogOpen(open);
    };

    if (loading) {
        return (
            <StatusCard
                description={t('settings.provider.loadingDescription')}
                icon={<Loader2 className="text-muted-foreground size-16 animate-spin" />}
                title={t('settings.provider.loadingTitle')}
            />
        );
    }

    if (error) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertTitle>{t('settings.provider.loadingError')}</AlertTitle>
                <AlertDescription>{error.message}</AlertDescription>
            </Alert>
        );
    }

    const providers = data?.settingsProviders?.models
        ? Object.keys(data?.settingsProviders.models).filter((key) => key !== '__typename')
        : [];

    const mutationError = createError || updateError || deleteError || testError || agentTestError || submitError;

    return (
        <>
            <div className="flex flex-col gap-4">
                <div className="flex flex-col gap-2">
                    <h2 className="flex items-center gap-2 text-lg font-semibold">
                        <Cpu className="text-muted-foreground size-5" />
                        {isNew ? t('settings.provider.newTitle') : t('settings.provider.settingsTitle')}
                    </h2>

                    <div className="text-muted-foreground">
                        {isNew ? t('settings.provider.configureNew') : t('settings.provider.updateDescription')}
                    </div>
                </div>

                <Form {...form}>
                    <form
                        className="flex flex-col gap-6"
                        id="provider-form"
                        onSubmit={handleFormSubmit(handleSubmit)}
                    >
                        {/* Error Alert */}
                        {mutationError && (
                            <Alert variant="destructive">
                                <AlertCircle className="size-4" />
                                <AlertTitle>{t('common.error')}</AlertTitle>
                                <AlertDescription>
                                    {mutationError instanceof Error ? (
                                        mutationError.message
                                    ) : (
                                        <div className="whitespace-pre-line">{mutationError}</div>
                                    )}
                                </AlertDescription>
                            </Alert>
                        )}

                        {/* Form fields */}
                        <FormComboboxItem
                            allowCustom={false}
                            control={control}
                            description={t('settings.provider.typeDescription')}
                            disabled={isLoading || !!selectedType}
                            label={t('settings.provider.type')}
                            name="type"
                            options={providers}
                            placeholder={t('settings.provider.selectType')}
                        />

                        <FormInputStringItem
                            control={control}
                            description={t('settings.provider.nameDescription')}
                            disabled={isLoading}
                            label={t('settings.provider.name')}
                            name="name"
                            placeholder={t('settings.provider.namePlaceholder')}
                        />

                        {/* Agents Configuration Section */}
                        <div className="flex flex-col gap-4">
                            <div>
                                <h3 className="text-lg font-medium">{t('settings.provider.agentConfigurations')}</h3>
                                <p className="text-muted-foreground text-sm">
                                    {t('settings.provider.agentConfigurationsDescription')}
                                </p>
                            </div>

                            <Accordion
                                className="w-full"
                                type="multiple"
                            >
                                {agentTypes.map((agentKey) => (
                                    <AccordionItem
                                        key={agentKey}
                                        value={agentKey}
                                    >
                                        <AccordionTrigger className="group text-left hover:no-underline">
                                            <div className="flex w-full items-center justify-between gap-2">
                                                <span className="group-hover:underline">
                                                    {translateAgentName(agentKey, t)}
                                                </span>
                                                <span
                                                    className={cn(
                                                        'hover:bg-accent hover:text-accent-foreground mr-2 flex items-center gap-1 rounded border px-2 py-1 text-xs',
                                                        (isTestLoading || isAgentTestLoading) &&
                                                            'pointer-events-none cursor-not-allowed opacity-50',
                                                    )}
                                                    onClick={(event) => {
                                                        if (isTestLoading || isAgentTestLoading) {
                                                            return;
                                                        }

                                                        event.stopPropagation();
                                                        handleTestAgent(agentKey);
                                                    }}
                                                >
                                                    {isAgentTestLoading && currentAgentKey === agentKey ? (
                                                        <Loader2 className="size-4 animate-spin" />
                                                    ) : (
                                                        <Play className="size-4" />
                                                    )}
                                                    <span className="no-underline! hover:no-underline!">
                                                        {isAgentTestLoading && currentAgentKey === agentKey
                                                            ? t('common.testing')
                                                            : t('common.test')}
                                                    </span>
                                                </span>
                                            </div>
                                        </AccordionTrigger>
                                        <AccordionContent className="flex flex-col gap-4 pt-4">
                                            <div className="grid grid-cols-1 gap-4 p-px md:grid-cols-2">
                                                {/* Model field */}
                                                <FormModelComboboxItem
                                                    control={control}
                                                    disabled={isLoading}
                                                    label={t('settings.provider.model')}
                                                    name={`agents.${agentKey}.model`}
                                                    onOptionSelect={(option) => {
                                                        {
                                                            /* Update price fields */
                                                        }

                                                        const price = option?.price;

                                                        setValue(
                                                            `agents.${agentKey}.price.input` as const,
                                                            price?.input ?? null,
                                                        );
                                                        setValue(
                                                            `agents.${agentKey}.price.output` as const,
                                                            price?.output ?? null,
                                                        );
                                                        setValue(
                                                            `agents.${agentKey}.price.cacheRead` as const,
                                                            price?.cacheRead ?? null,
                                                        );
                                                        setValue(
                                                            `agents.${agentKey}.price.cacheWrite` as const,
                                                            price?.cacheWrite ?? null,
                                                        );
                                                    }}
                                                    options={availableModels}
                                                    placeholder={t('settings.provider.modelPlaceholder')}
                                                />

                                                {/* Temperature field */}
                                                <FormInputNumberItem
                                                    control={control}
                                                    disabled={isLoading}
                                                    label={t('settings.provider.temperature')}
                                                    max="2"
                                                    min="0"
                                                    name={`agents.${agentKey}.temperature`}
                                                    placeholder="0.7"
                                                    step="0.1"
                                                />

                                                {/* Max Tokens field */}
                                                <FormInputNumberItem
                                                    control={control}
                                                    disabled={isLoading}
                                                    label={t('settings.provider.maxTokens')}
                                                    min="1"
                                                    name={`agents.${agentKey}.maxTokens`}
                                                    placeholder="1000"
                                                    valueType="integer"
                                                />

                                                {/* Top P field */}
                                                <FormInputNumberItem
                                                    control={control}
                                                    disabled={isLoading}
                                                    label={t('settings.provider.topP')}
                                                    max="1"
                                                    min="0"
                                                    name={`agents.${agentKey}.topP`}
                                                    placeholder="0.9"
                                                    step="0.01"
                                                />

                                                {/* Top K field */}
                                                <FormInputNumberItem
                                                    control={control}
                                                    disabled={isLoading}
                                                    label={t('settings.provider.topK')}
                                                    min="1"
                                                    name={`agents.${agentKey}.topK`}
                                                    placeholder="40"
                                                    valueType="integer"
                                                />

                                                {/* Min Length field */}
                                                <FormInputNumberItem
                                                    control={control}
                                                    disabled={isLoading}
                                                    label={t('settings.provider.minLength')}
                                                    min="0"
                                                    name={`agents.${agentKey}.minLength`}
                                                    placeholder="0"
                                                    valueType="integer"
                                                />

                                                {/* Max Length field */}
                                                <FormInputNumberItem
                                                    control={control}
                                                    disabled={isLoading}
                                                    label={t('settings.provider.maxLength')}
                                                    min="1"
                                                    name={`agents.${agentKey}.maxLength`}
                                                    placeholder="2000"
                                                    valueType="integer"
                                                />

                                                {/* Repetition Penalty field */}
                                                <FormInputNumberItem
                                                    control={control}
                                                    disabled={isLoading}
                                                    label={t('settings.provider.repetitionPenalty')}
                                                    max="2"
                                                    min="0"
                                                    name={`agents.${agentKey}.repetitionPenalty`}
                                                    placeholder="1.0"
                                                    step="0.01"
                                                />

                                                {/* Frequency Penalty field */}
                                                <FormInputNumberItem
                                                    control={control}
                                                    disabled={isLoading}
                                                    label={t('settings.provider.frequencyPenalty')}
                                                    max="2"
                                                    min="0"
                                                    name={`agents.${agentKey}.frequencyPenalty`}
                                                    placeholder="0.0"
                                                    step="0.01"
                                                />

                                                {/* Presence Penalty field */}
                                                <FormInputNumberItem
                                                    control={control}
                                                    disabled={isLoading}
                                                    label={t('settings.provider.presencePenalty')}
                                                    max="2"
                                                    min="0"
                                                    name={`agents.${agentKey}.presencePenalty`}
                                                    placeholder="0.0"
                                                    step="0.01"
                                                />
                                            </div>

                                            {/* Reasoning Configuration */}
                                            <div className="col-span-full p-px">
                                                <div className="mt-6 flex flex-col gap-4">
                                                    <h4 className="text-sm font-medium">
                                                        {t('settings.provider.reasoningConfiguration')}
                                                    </h4>
                                                    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                                                        {/* Reasoning Effort field */}
                                                        <FormField
                                                            control={control}
                                                            name={`agents.${agentKey}.reasoning.effort`}
                                                            render={({ field }) => (
                                                                <FormItem>
                                                                    <FormLabel>
                                                                        {t('settings.provider.reasoningEffort')}
                                                                    </FormLabel>
                                                                    <Select
                                                                        defaultValue={field.value ?? 'none'}
                                                                        disabled={isLoading}
                                                                        onValueChange={(value) =>
                                                                            field.onChange(
                                                                                value !== 'none' ? value : null,
                                                                            )
                                                                        }
                                                                    >
                                                                        <FormControl>
                                                                            <SelectTrigger>
                                                                                <SelectValue
                                                                                    placeholder={t(
                                                                                        'settings.provider.selectEffort',
                                                                                    )}
                                                                                />
                                                                            </SelectTrigger>
                                                                        </FormControl>
                                                                        <SelectContent>
                                                                            <SelectItem value="none">
                                                                                {t('common.notSelected')}
                                                                            </SelectItem>
                                                                            <SelectItem value={ReasoningEffort.Low}>
                                                                                {t('common.low')}
                                                                            </SelectItem>
                                                                            <SelectItem value={ReasoningEffort.Medium}>
                                                                                {t('common.medium')}
                                                                            </SelectItem>
                                                                            <SelectItem value={ReasoningEffort.High}>
                                                                                {t('common.high')}
                                                                            </SelectItem>
                                                                        </SelectContent>
                                                                    </Select>
                                                                    <FormMessage />
                                                                </FormItem>
                                                            )}
                                                        />

                                                        {/* Reasoning Max Tokens field */}
                                                        <FormInputNumberItem
                                                            control={control}
                                                            disabled={isLoading}
                                                            label={t('settings.provider.reasoningMaxTokens')}
                                                            min="1"
                                                            name={`agents.${agentKey}.reasoning.maxTokens`}
                                                            placeholder="1000"
                                                            valueType="integer"
                                                        />
                                                    </div>
                                                </div>
                                            </div>

                                            {/* Price Configuration */}
                                            <div className="col-span-full p-px">
                                                <div className="mt-6 flex flex-col gap-4">
                                                    <h4 className="text-sm font-medium">
                                                        {t('settings.provider.priceConfiguration')}
                                                    </h4>
                                                    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                                                        {/* Price Input field */}
                                                        <FormInputNumberItem
                                                            control={control}
                                                            description={t('settings.provider.inputPriceDescription')}
                                                            disabled={isLoading}
                                                            label={t('settings.provider.inputPrice')}
                                                            min="0"
                                                            name={`agents.${agentKey}.price.input`}
                                                            placeholder="0.001"
                                                            step="0.000001"
                                                        />

                                                        {/* Price Output field */}
                                                        <FormInputNumberItem
                                                            control={control}
                                                            description={t('settings.provider.outputPriceDescription')}
                                                            disabled={isLoading}
                                                            label={t('settings.provider.outputPrice')}
                                                            min="0"
                                                            name={`agents.${agentKey}.price.output`}
                                                            placeholder="0.002"
                                                            step="0.000001"
                                                        />

                                                        {/* Cache Read Price field */}
                                                        <FormInputNumberItem
                                                            control={control}
                                                            description={t(
                                                                'settings.provider.cacheReadPriceDescription',
                                                            )}
                                                            disabled={isLoading}
                                                            label={t('settings.provider.cacheReadPrice')}
                                                            min="0"
                                                            name={`agents.${agentKey}.price.cacheRead`}
                                                            placeholder="0.0001"
                                                            step="0.000001"
                                                        />

                                                        {/* Cache Write Price field */}
                                                        <FormInputNumberItem
                                                            control={control}
                                                            description={t(
                                                                'settings.provider.cacheWritePriceDescription',
                                                            )}
                                                            disabled={isLoading}
                                                            label={t('settings.provider.cacheWritePrice')}
                                                            min="0"
                                                            name={`agents.${agentKey}.price.cacheWrite`}
                                                            placeholder="0.00015"
                                                            step="0.000001"
                                                        />
                                                    </div>
                                                </div>
                                            </div>
                                        </AccordionContent>
                                    </AccordionItem>
                                ))}
                            </Accordion>
                        </div>
                    </form>
                </Form>
            </div>

            {/* Sticky buttons at bottom */}
            <div className="bg-background sticky -bottom-4 -mx-4 mt-4 -mb-4 flex items-center border-t p-4 shadow-lg">
                <div className="flex gap-2">
                    {/* Delete button - only show when editing existing provider */}
                    {!isNew && (
                        <Button
                            disabled={isLoading}
                            onClick={handleDelete}
                            type="button"
                            variant="destructive"
                        >
                            {isDeleteLoading ? (
                                <Loader2 className="size-4 animate-spin" />
                            ) : (
                                <Trash2 className="size-4" />
                            )}
                            {isDeleteLoading ? t('common.deleting') : t('common.delete')}
                        </Button>
                    )}
                    <Button
                        disabled={isLoading || isTestLoading || isAgentTestLoading}
                        onClick={() => handleTest()}
                        type="button"
                        variant="outline"
                    >
                        {isTestLoading ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
                        {isTestLoading ? t('common.testing') : t('common.test')}
                    </Button>
                </div>

                <div className="ml-auto flex gap-2">
                    <Button
                        disabled={isLoading}
                        onClick={handleBack}
                        type="button"
                        variant="outline"
                    >
                        {t('common.cancel')}
                    </Button>
                    <FormSubmitButton
                        form="provider-form"
                        icon={<Save className="size-4" />}
                        loading={isLoading}
                        variant="secondary"
                    >
                        {isLoading
                            ? t('common.saving')
                            : isNew
                              ? t('settings.createProvider')
                              : t('settings.updateProvider')}
                    </FormSubmitButton>
                </div>
            </div>

            <TestResultsDialog
                handleOpenChange={setIsTestDialogOpen}
                isOpen={isTestDialogOpen}
                results={testResults}
            />

            <ConfirmationDialog
                cancelText={t('common.cancel')}
                confirmText={t('common.delete')}
                description={t('settings.provider.deleteDescription', { name: providerName ?? '' })}
                handleConfirm={handleConfirmDelete}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
                itemName={providerName}
                itemType="provider"
                title={t('settings.provider.deleteTitle')}
            />

            <ConfirmationDialog
                cancelText={t('common.stay')}
                confirmIcon={undefined}
                confirmText={t('common.leave')}
                confirmVariant="destructive"
                description={t('settings.unsavedDescription')}
                handleConfirm={handleConfirmLeave}
                handleOpenChange={handleLeaveDialogOpenChange}
                isOpen={isLeaveDialogOpen}
                title={t('settings.unsavedTitle')}
            />
        </>
    );
}

export default SettingsProvider;
