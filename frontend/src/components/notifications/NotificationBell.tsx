import { Bell } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { EmptyState, IconButton, LoadingState } from "../common";
import {
  formatNotificationTime,
  getNotificationLink,
  getUnreadNotificationCount,
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
  type NotificationItem,
} from "../../services/notificationService";
import { toast } from "../../services/toast";

export function NotificationBell() {
  const [isOpen, setIsOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  const [items, setItems] = useState<NotificationItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isMarkingAll, setIsMarkingAll] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  const fetchUnreadCount = async () => {
    try {
      const response = await getUnreadNotificationCount();
      setUnreadCount(response.count);
    } catch {
      // silent for badge
    }
  };

  const fetchRecent = async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      const response = await listNotifications({ page: 1, pageSize: 5 });
      setItems(response.items);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Bildirimler yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchUnreadCount();
    const interval = window.setInterval(() => void fetchUnreadCount(), 60000);
    const handleFocus = () => void fetchUnreadCount();
    window.addEventListener("focus", handleFocus);
    return () => {
      window.clearInterval(interval);
      window.removeEventListener("focus", handleFocus);
    };
  }, []);

  useEffect(() => {
    if (!isOpen) return;
    void fetchRecent();
  }, [isOpen]);

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      if (!wrapperRef.current?.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    window.addEventListener("pointerdown", handlePointerDown);
    return () => window.removeEventListener("pointerdown", handlePointerDown);
  }, []);

  const handleMarkAllRead = async () => {
    if (unreadCount === 0 || isMarkingAll) return;
    setIsMarkingAll(true);
    try {
      await markAllNotificationsRead();
      setUnreadCount(0);
      setItems((current) => current.map((item) => ({ ...item, isRead: true })));
      toast.success("Tüm bildirimler okundu olarak işaretlendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Bildirimler okundu olarak işaretlenemedi.");
    } finally {
      setIsMarkingAll(false);
    }
  };

  const handleNotificationClick = async (notification: NotificationItem) => {
    try {
      if (!notification.isRead) {
        await markNotificationRead(notification.id);
        setUnreadCount((current) => Math.max(0, current - 1));
        setItems((current) => current.map((item) => (item.id === notification.id ? { ...item, isRead: true } : item)));
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Bildirim okundu olarak işaretlenemedi.");
    }

    const link = getNotificationLink(notification);
    setIsOpen(false);
    if (link) {
      navigate(link);
    }
  };

  return (
    <div ref={wrapperRef} className="relative shrink-0">
      <div className="relative">
        <IconButton aria-label="Bildirimler" onClick={() => setIsOpen((current) => !current)}>
          <Bell className="h-4 w-4" />
        </IconButton>
        {unreadCount > 0 ? (
          <span className="pointer-events-none absolute -right-1 -top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-violet-brand px-1 text-[10px] font-bold leading-none text-white ring-2 ring-white">
            {unreadCount > 99 ? "99+" : unreadCount}
          </span>
        ) : null}
      </div>

      {isOpen ? (
        <div className="absolute right-0 top-full z-50 mt-2 w-[min(100vw-2rem,380px)] overflow-hidden rounded-xl border border-line bg-white shadow-panel">
          <div className="flex items-center justify-between gap-2 border-b border-line px-4 py-3">
            <h2 className="text-sm font-bold text-ink">Bildirimler</h2>
            <div className="flex items-center gap-2">
              {unreadCount > 0 ? (
                <>
                  <span className="text-xs font-semibold text-violet-brand">{unreadCount} okunmamış</span>
                  <button
                    type="button"
                    className="text-xs font-semibold text-slate-600 hover:text-violet-brand disabled:opacity-50"
                    disabled={isMarkingAll}
                    onClick={() => void handleMarkAllRead()}
                  >
                    {isMarkingAll ? "İşaretleniyor..." : "Tümünü okundu işaretle"}
                  </button>
                </>
              ) : null}
            </div>
          </div>

          {isLoading ? (
            <div className="p-4">
              <LoadingState label="Bildirimler yükleniyor..." />
            </div>
          ) : loadError ? (
            <div className="p-4">
              <EmptyState title="Bildirimler yüklenemedi" description={loadError} />
            </div>
          ) : items.length === 0 ? (
            <div className="p-4">
              <EmptyState title="Henüz bildirimin yok" description="Yeni olaylar burada görünecek." />
            </div>
          ) : (
            <div className="max-h-96 divide-y divide-line overflow-y-auto">
              {items.map((notification) => (
                <button
                  key={notification.id}
                  type="button"
                  className={`w-full px-4 py-3 text-left transition hover:bg-violet-soft/40 ${notification.isRead ? "bg-white" : "bg-violet-soft/30"}`}
                  onClick={() => void handleNotificationClick(notification)}
                >
                  <p className="text-sm font-bold text-ink">{notification.title}</p>
                  <p className="mt-1 line-clamp-2 text-sm text-slate-600">{notification.message}</p>
                  <p className="mt-2 text-xs font-medium text-slate-400">{formatNotificationTime(notification.createdAt)}</p>
                </button>
              ))}
            </div>
          )}

          <div className="border-t border-line px-4 py-3">
            <Link className="text-sm font-semibold text-violet-brand hover:underline" to="/notifications" onClick={() => setIsOpen(false)}>
              Tümünü Gör
            </Link>
          </div>
        </div>
      ) : null}
    </div>
  );
}
