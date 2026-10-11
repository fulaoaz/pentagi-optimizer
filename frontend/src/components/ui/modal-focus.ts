import * as React from 'react';

type FocusTargetRef = React.RefObject<HTMLElement | null>;

export const ModalFocusContext = React.createContext<FocusTargetRef | null>(null);

export function isRestorableFocusTarget(target: Element | null): target is HTMLElement {
    return (
        target instanceof HTMLElement &&
        target !== document.body &&
        !target.closest('[data-radix-focus-guard], [role="dialog"], [role="menu"]')
    );
}

export function useModalFocusTarget() {
    const targetRef = React.useRef<HTMLElement | null>(null);

    React.useEffect(() => {
        const rememberTarget = (event: Event) => {
            const target = getFocusTarget(event.target);

            if (target) {
                targetRef.current = target;
            }
        };

        document.addEventListener('contextmenu', rememberTarget, true);
        document.addEventListener('focusin', rememberTarget, true);
        document.addEventListener('pointerdown', rememberTarget, true);

        return () => {
            document.removeEventListener('contextmenu', rememberTarget, true);
            document.removeEventListener('focusin', rememberTarget, true);
            document.removeEventListener('pointerdown', rememberTarget, true);
        };
    }, []);

    return targetRef;
}

function getFocusTarget(target: EventTarget | null) {
    if (!(target instanceof HTMLElement)) {
        return null;
    }

    const candidate = target.closest<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex], [data-slot$="-trigger"]',
    );

    return isRestorableFocusTarget(candidate) ? candidate : null;
}
