import type { ComponentType } from "react";

export type AppIcon = ComponentType<{ className?: string }>;

export type FileBadgeType = "pdf" | "doc" | "sheet";

export type NavGroup = {
  label?: string;
  items: {
    label: string;
    path: string;
    end?: boolean;
    primary?: boolean;
    roles?: Array<"owner" | "admin" | "member">;
    icon: AppIcon;
  }[];
};
