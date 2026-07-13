import type { OnboardingStepId } from "./onboardingConfig";

type OnboardingState = {
  dismissed: boolean;
  manualCompleted: OnboardingStepId[];
};

function storageKey(userId: string, tenantId: string) {
  return `fifriz_onboarding_${userId}_${tenantId}`;
}

function readJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return fallback;
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

function writeJSON<T>(key: string, value: T) {
  localStorage.setItem(key, JSON.stringify(value));
}

function readState(userId: string, tenantId: string): OnboardingState {
  return readJSON<OnboardingState>(storageKey(userId, tenantId), {
    dismissed: false,
    manualCompleted: [],
  });
}

function writeState(userId: string, tenantId: string, state: OnboardingState) {
  writeJSON(storageKey(userId, tenantId), state);
}

export function isOnboardingDismissed(userId: string, tenantId: string) {
  return readState(userId, tenantId).dismissed;
}

export function dismissOnboarding(userId: string, tenantId: string) {
  const state = readState(userId, tenantId);
  writeState(userId, tenantId, { ...state, dismissed: true });
}

export function isStepManuallyCompleted(userId: string, tenantId: string, stepId: OnboardingStepId) {
  return readState(userId, tenantId).manualCompleted.includes(stepId);
}

export function markOnboardingStep(userId: string, tenantId: string, stepId: OnboardingStepId) {
  const state = readState(userId, tenantId);
  if (state.manualCompleted.includes(stepId)) return;
  writeState(userId, tenantId, {
    ...state,
    manualCompleted: [...state.manualCompleted, stepId],
  });
}

export function notifyOnboardingStep(stepId: OnboardingStepId) {
  window.dispatchEvent(new CustomEvent("fifriz:onboarding-step", { detail: { stepId } }));
}

export function resetOnboardingDismissed(userId: string, tenantId: string) {
  const state = readState(userId, tenantId);
  writeState(userId, tenantId, { ...state, dismissed: false });
}
