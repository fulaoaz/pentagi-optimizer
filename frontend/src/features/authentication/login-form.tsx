import { zodResolver } from '@hookform/resolvers/zod';
import { Languages } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { z } from 'zod';

import type { Locale, Translate } from '@/lib/i18n';
import type { OAuthProvider } from '@/providers/user-provider';

import Github from '@/components/icons/github';
import Google from '@/components/icons/google';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { FormSubmitButton } from '@/components/ui/form-submit-button';
import { Input } from '@/components/ui/input';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useLocale } from '@/hooks/use-locale';
import { localeNames, locales } from '@/lib/i18n';
import { useUser } from '@/providers/user-provider';

import { PasswordChangeForm } from './password-change-form';

// Built inside the component so validation messages follow the active locale;
// a module-level schema would capture whichever locale was loaded first.
const buildFormSchema = (t: Translate) =>
    z.object({
        mail: z
            .string()
            .min(1, {
                message: t('auth.loginRequired'),
            })
            .refine(
                (value) =>
                    z.string().email().safeParse(value).success || ['admin', 'demo'].includes(value.toLowerCase()),
                {
                    message: t('auth.invalidLogin'),
                },
            ),
        password: z.string().min(1, {
            message: t('auth.passwordRequired'),
        }),
    });

interface AuthProviderAction {
    icon: React.ReactNode;
    id: OAuthProvider;
    nameKey: string;
}

const providerActions: AuthProviderAction[] = [
    {
        icon: <Google className="size-5" />,
        id: 'google',
        nameKey: 'auth.continueWithGoogle',
    },
    {
        icon: <Github className="size-5" />,
        id: 'github',
        nameKey: 'auth.continueWithGithub',
    },
];

interface LoginFormProps {
    providers: string[]; // OAuth providers: ['google', 'github']
    returnUrl?: string;
}

function LoginForm({ providers, returnUrl = '/flows/new' }: LoginFormProps) {
    const { locale, setLocale, t } = useLocale();
    const formSchema = useMemo(() => buildFormSchema(t), [t]);
    const errorMessage = t('auth.invalidLoginOrPassword');
    const errorProviderMessage = t('auth.providerFailed');
    const form = useForm<z.infer<typeof formSchema>>({
        defaultValues: {
            mail: '',
            password: '',
        },
        resolver: zodResolver(formSchema),
    });
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<null | string>(null);
    const [passwordChangeRequired, setPasswordChangeRequired] = useState(false);
    const navigate = useNavigate();
    const { authInfo, isAuthenticated, login, loginWithOAuth, setAuth } = useUser();

    const handleSubmit = async (values: z.infer<typeof formSchema>) => {
        setError(null);

        try {
            const result = await login(values);

            if (!result.success) {
                setError(result.error || errorMessage);

                return;
            }

            if (result.passwordChangeRequired) {
                setPasswordChangeRequired(true);

                return;
            }

            navigate(returnUrl);
        } catch {
            setError(errorMessage);
        }
    };

    const handleProviderLogin = async (provider: OAuthProvider) => {
        setError(null);
        setIsSubmitting(true);

        try {
            const result = await loginWithOAuth(provider);

            if (!result.success) {
                setError(result.error || errorProviderMessage);

                return;
            }

            navigate(returnUrl);
        } catch (error) {
            setError(error instanceof Error ? error.message : errorMessage);
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleSkipPasswordChange = () => {
        navigate(returnUrl);
    };

    const handlePasswordChangeSuccess = () => {
        if (authInfo?.user) {
            const updatedAuthData = {
                ...authInfo,
                user: {
                    ...authInfo.user,
                    password_change_required: false,
                },
            };

            setAuth(updatedAuthData);
            navigate(returnUrl);
        }
    };

    // If password change is required, show password change form.
    // Also check isAuthenticated() to ensure the user has a valid session.
    // If the session expired and user refreshed the page, the old authInfo may still
    // be in memory (race condition between clearAuth() and navigate()), but we must
    // NOT show the password change form because:
    //   1. The API endpoint /user/password requires authentication (returns 403 if not)
    //   2. The user must first re-login to establish a new valid session
    // Also check authInfo directly to handle page refresh scenarios where passwordChangeRequired
    // local state is lost but authInfo.user.password_change_required is still true.
    const shouldShowPasswordChange =
        (passwordChangeRequired || authInfo?.user?.password_change_required) &&
        authInfo?.user?.type === 'local' &&
        isAuthenticated();

    if (shouldShowPasswordChange) {
        return (
            <div className="mx-auto flex w-[350px] flex-col gap-6">
                <h1 className="text-center text-3xl font-bold">{t('auth.updatePassword')}</h1>
                <p className="text-muted-foreground text-center text-sm">{t('auth.passwordChangeHint')}</p>
                <PasswordChangeForm
                    isModal={false}
                    onSkip={handleSkipPasswordChange}
                    onSuccess={handlePasswordChangeSuccess}
                    showSkip={true}
                />
            </div>
        );
    }

    return (
        <Form {...form}>
            <form
                className="mx-auto grid w-[350px] gap-8"
                onSubmit={form.handleSubmit(handleSubmit)}
            >
                <div className="space-y-3">
                    <h1 className="text-center text-3xl font-bold">PentAGI</h1>
                    <div className="flex items-center justify-center gap-2 text-sm">
                        <Languages
                            aria-hidden="true"
                            className="text-muted-foreground size-4"
                        />
                        <span className="text-muted-foreground">{t('settings.language')}</span>
                        <Tabs
                            onValueChange={(value) => setLocale(value as Locale)}
                            value={locale}
                        >
                            <TabsList className="h-8 p-0.5">
                                {locales.map((value) => (
                                    <TabsTrigger
                                        className="h-7 px-2 text-xs"
                                        key={value}
                                        value={value}
                                    >
                                        {localeNames[value]}
                                    </TabsTrigger>
                                ))}
                            </TabsList>
                        </Tabs>
                    </div>
                </div>

                {providers?.length > 0 && (
                    <>
                        <div className="flex flex-col gap-4">
                            {providerActions
                                .filter((provider) => providers.includes(provider.id))
                                .map((provider) => (
                                    <Button
                                        disabled={isSubmitting || form.formState.isSubmitting}
                                        key={provider.id}
                                        onClick={() => handleProviderLogin(provider.id)}
                                        type="button"
                                        variant="secondary"
                                    >
                                        {provider.icon}
                                        {t(provider.nameKey)}
                                    </Button>
                                ))}
                        </div>

                        <div className="relative -mb-4">
                            <div className="absolute inset-0 flex items-center">
                                <div className="w-full border-t border-gray-300" />
                            </div>
                            <div className="relative flex justify-center text-sm">
                                <span className="bg-background px-2">{t('auth.or')}</span>
                            </div>
                        </div>
                    </>
                )}

                <div className="flex flex-col gap-4">
                    <FormField
                        control={form.control}
                        name="mail"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>{t('auth.login')}</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        autoFocus
                                        placeholder={t('auth.enterEmail')}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="password"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>{t('auth.password')}</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        placeholder={t('auth.enterPassword')}
                                        type="password"
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormSubmitButton className="w-full">
                        <span>{t('auth.signIn')}</span>
                    </FormSubmitButton>

                    {error && <FormMessage>{error}</FormMessage>}
                </div>
            </form>
        </Form>
    );
}

export default LoginForm;
