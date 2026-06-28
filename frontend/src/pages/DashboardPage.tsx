import { useState, useEffect, useCallback } from 'react';
import QRCode from 'qrcode';
import { useAuth } from '@/context/AuthContext';
import {
  getVlessClientsByUserId,
  getHysteriaClientsByUserId,
  getXrayLink,
  getHysteriaLink,
  updateVlessClientAlias,
  updateHysteriaClientAlias,
  deleteVlessClient,
  deleteHysteriaClient,
} from '@/api/client';
import type { ClientPublicOut } from '@/types';

type VpnType = 'vless' | 'hysteria';

export default function DashboardPage() {
  const { user, logout } = useAuth();

  const [vlessClients, setVlessClients] = useState<ClientPublicOut[]>([]);
  const [hysteriaClients, setHysteriaClients] = useState<ClientPublicOut[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Link state
  const [activeClientId, setActiveClientId] = useState<number | null>(null);
  const [activeClientType, setActiveClientType] = useState<VpnType | null>(null);
  const [link, setLink] = useState('');
  const [linkLoading, setLinkLoading] = useState(false);
  const [linkError, setLinkError] = useState('');
  const [copied, setCopied] = useState(false);
  const [qrDataUrl, setQrDataUrl] = useState('');

  // Edit alias state
  const [editingClientId, setEditingClientId] = useState<number | null>(null);
  const [editingClientType, setEditingClientType] = useState<VpnType | null>(null);
  const [editAlias, setEditAlias] = useState('');
  const [editLoading, setEditLoading] = useState(false);
  const [editError, setEditError] = useState('');

  // Delete confirmation state
  const [deleteConfirmId, setDeleteConfirmId] = useState<number | null>(null);
  const [deleteConfirmType, setDeleteConfirmType] = useState<VpnType | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [deleteError, setDeleteError] = useState('');

  const loadClients = useCallback(() => {
    if (!user) return;
    setLoading(true);
    setError('');
    Promise.all([
      getVlessClientsByUserId(user.id).catch((err) => {
        setError(err.message);
        return [] as ClientPublicOut[];
      }),
      getHysteriaClientsByUserId(user.id).catch((err) => {
        setError(err.message);
        return [] as ClientPublicOut[];
      }),
    ])
      .then(([vless, hysteria]) => {
        setVlessClients(vless);
        setHysteriaClients(hysteria);
      })
      .finally(() => setLoading(false));
  }, [user]);

  useEffect(() => {
    loadClients();
  }, [loadClients]);

  // --- Link handlers ---

  const handleGetLink = async (clientId: number, clientType: VpnType) => {
    if (activeClientId === clientId && activeClientType === clientType && link) {
      setActiveClientId(null);
      setActiveClientType(null);
      setLink('');
      setQrDataUrl('');
      return;
    }
    setActiveClientId(clientId);
    setActiveClientType(clientType);
    setLink('');
    setLinkError('');
    setCopied(false);
    setQrDataUrl('');
    setLinkLoading(true);
    try {
      const result =
        clientType === 'vless'
          ? await getXrayLink(String(clientId))
          : await getHysteriaLink(clientId);
      setLink(result.connection_link);
      const canQr =
        result.connection_link.startsWith('vless://') ||
        result.connection_link.startsWith('hysteria2://');
      if (canQr) {
        const qr = await QRCode.toDataURL(result.connection_link, {
          width: 256,
          margin: 2,
          color: { dark: '#000000', light: '#ffffff' },
        });
        setQrDataUrl(qr);
      }
    } catch (err: any) {
      setLinkError(err.message || 'Не удалось получить ссылку');
    } finally {
      setLinkLoading(false);
    }
  };

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(link);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      const textarea = document.createElement('textarea');
      textarea.value = link;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  // --- Alias edit handlers ---

  const handleStartEdit = (client: ClientPublicOut, clientType: VpnType) => {
    setEditingClientId(client.id);
    setEditingClientType(clientType);
    setEditAlias(client.alias);
    setEditError('');
  };

  const handleCancelEdit = () => {
    setEditingClientId(null);
    setEditingClientType(null);
    setEditAlias('');
    setEditError('');
  };

  const handleSaveAlias = async (clientId: number, clientType: VpnType) => {
    if (!editAlias.trim()) return;
    setEditLoading(true);
    setEditError('');
    try {
      const updated =
        clientType === 'vless'
          ? await updateVlessClientAlias(clientId, { new_alias: editAlias.trim() })
          : await updateHysteriaClientAlias(clientId, { new_alias: editAlias.trim() });
      if (clientType === 'vless') {
        setVlessClients((prev) => prev.map((c) => (c.id === clientId ? updated : c)));
      } else {
        setHysteriaClients((prev) => prev.map((c) => (c.id === clientId ? updated : c)));
      }
      setEditingClientId(null);
      setEditingClientType(null);
    } catch (err: any) {
      setEditError(err.message || 'Ошибка сохранения');
    } finally {
      setEditLoading(false);
    }
  };

  // --- Delete handlers ---

  const handleDeleteClick = (clientId: number, clientType: VpnType) => {
    setDeleteConfirmId(clientId);
    setDeleteConfirmType(clientType);
    setDeleteError('');
  };

  const handleCancelDelete = () => {
    setDeleteConfirmId(null);
    setDeleteConfirmType(null);
    setDeleteError('');
  };

  const handleConfirmDelete = async () => {
    if (!deleteConfirmId || !deleteConfirmType) return;
    setDeleteLoading(true);
    setDeleteError('');
    try {
      if (deleteConfirmType === 'vless') {
        await deleteVlessClient(deleteConfirmId);
        setVlessClients((prev) => prev.filter((c) => c.id !== deleteConfirmId));
      } else {
        await deleteHysteriaClient(deleteConfirmId);
        setHysteriaClients((prev) => prev.filter((c) => c.id !== deleteConfirmId));
      }
      if (activeClientId === deleteConfirmId) {
        setActiveClientId(null);
        setActiveClientType(null);
        setLink('');
        setQrDataUrl('');
      }
      setDeleteConfirmId(null);
      setDeleteConfirmType(null);
    } catch (err: any) {
      setDeleteError(err.message || 'Ошибка удаления');
    } finally {
      setDeleteLoading(false);
    }
  };

  // --- Helpers ---

  const formatDate = (iso: string) => {
    return new Date(iso).toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
    });
  };

  const statusLabel = (status: string) => {
    switch (status) {
      case 'active': return 'Активен';
      case 'disabled': return 'Отключен';
      default: return status;
    }
  };

  const statusColor = (status: string) => {
    switch (status) {
      case 'active': return 'text-green-400';
      case 'disabled': return 'text-red-400';
      default: return 'text-gray-400';
    }
  };

  const isLinkActive = (clientId: number, clientType: VpnType) =>
    activeClientId === clientId && activeClientType === clientType;

  const isEditing = (clientId: number, clientType: VpnType) =>
    editingClientId === clientId && editingClientType === clientType;

  // --- Client row renderer ---

  const renderClientCard = (client: ClientPublicOut, clientType: VpnType) => (
    <div
      key={`${clientType}-${client.id}`}
      className="bg-gray-800 border border-gray-700 rounded-xl p-4"
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center">
          {isEditing(client.id, clientType) ? (
            <div className="flex items-center gap-2">
              <input
                type="text"
                value={editAlias}
                onChange={(e) => setEditAlias(e.target.value)}
                className="px-3 py-1 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:border-blue-500 w-40"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleSaveAlias(client.id, clientType);
                  if (e.key === 'Escape') handleCancelEdit();
                }}
              />
              <button
                onClick={() => handleSaveAlias(client.id, clientType)}
                disabled={editLoading}
                className="px-3 py-1 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white text-sm rounded-lg transition"
              >
                {editLoading ? '...' : '✓'}
              </button>
              <button
                onClick={handleCancelEdit}
                className="px-3 py-1 bg-gray-600 hover:bg-gray-500 text-white text-sm rounded-lg transition"
              >
                ✕
              </button>
            </div>
          ) : (
            <>
              <span className="text-white font-medium">{client.alias}</span>
              <button
                onClick={() => handleStartEdit(client, clientType)}
                className="ml-2 text-gray-500 hover:text-gray-300 transition"
                title="Изменить алиас"
              >
                ✎
              </button>
            </>
          )}
          <span className={`ml-3 text-sm ${statusColor(client.status)}`}>
            {statusLabel(client.status)}
          </span>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-gray-500 text-sm">с {formatDate(client.created_at)}</span>
          <button
            onClick={() => handleDeleteClick(client.id, clientType)}
            className="text-red-500 hover:text-red-400 transition text-lg leading-none"
            title="Удалить клиента"
          >
            ✕
          </button>
          <button
            onClick={() => handleGetLink(client.id, clientType)}
            className="px-4 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded-lg transition"
          >
            {isLinkActive(client.id, clientType) && linkLoading ? 'Загрузка...' : 'Получить ссылку'}
          </button>
        </div>
      </div>

      {editError && isEditing(client.id, clientType) && (
        <p className="mt-2 text-red-400 text-xs">{editError}</p>
      )}

      {isLinkActive(client.id, clientType) && (
        <div className="mt-3 pt-3 border-t border-gray-700">
          {linkLoading ? (
            <p className="text-gray-400 text-sm">Получаем ссылку...</p>
          ) : linkError ? (
            <p className="text-red-400 text-sm">{linkError}</p>
          ) : link ? (
            <div>
              <label className="block text-gray-400 text-xs mb-1">Ссылка подключения:</label>
              <div className="flex gap-2">
                <input
                  readOnly
                  value={link}
                  className="flex-1 px-3 py-2 bg-gray-900 border border-gray-600 rounded-lg text-green-400 text-sm font-mono focus:outline-none"
                />
                <button
                  onClick={handleCopy}
                  className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white text-sm rounded-lg transition whitespace-nowrap"
                >
                  {copied ? 'Скопировано!' : 'Копировать'}
                </button>
              </div>
              {qrDataUrl && (
                <div className="mt-3 flex flex-col items-center">
                  <p className="text-gray-400 text-xs mb-2">Отсканируйте QR-код в приложении:</p>
                  <img
                    src={qrDataUrl}
                    alt="QR-код для подключения"
                    className="w-48 h-48 rounded-lg bg-white p-1"
                  />
                </div>
              )}
            </div>
          ) : null}
        </div>
      )}
    </div>
  );

  const renderClientList = (
    clients: ClientPublicOut[],
    clientType: VpnType,
    title: string,
    emptyText: string,
  ) => (
    <div className="mb-8">
      <h2 className="text-lg font-semibold text-white mb-4">{title}</h2>
      {clients.length === 0 ? (
        <div className="p-6 bg-gray-800 rounded-xl border border-gray-700 text-center">
          <p className="text-gray-400">{emptyText}</p>
          <a href="/invites/activate" className="text-blue-400 hover:underline text-sm mt-2 inline-block">
            Активируйте инвайт, чтобы создать клиента
          </a>
        </div>
      ) : (
        <div className="space-y-3">
          {clients.map((client) => renderClientCard(client, clientType))}
        </div>
      )}
    </div>
  );

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Delete confirmation modal */}
      {deleteConfirmId !== null && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
          <div className="bg-gray-800 border border-gray-600 rounded-xl p-6 max-w-sm w-full mx-4 shadow-2xl">
            <h3 className="text-white font-semibold text-lg mb-2">Подтверждение удаления</h3>
            <p className="text-gray-300 text-sm mb-1">
              Вы уверены, что хотите удалить клиента{deleteConfirmType === 'vless' ? ' VLESS' : ' Hysteria'}?
            </p>
            <p className="text-gray-400 text-xs mb-4">Это действие нельзя отменить.</p>
            {deleteError && (
              <p className="text-red-400 text-xs mb-3">{deleteError}</p>
            )}
            <div className="flex gap-3">
              <button
                onClick={handleCancelDelete}
                disabled={deleteLoading}
                className="flex-1 px-4 py-2 bg-gray-600 hover:bg-gray-500 disabled:opacity-50 text-white text-sm rounded-lg transition"
              >
                Отмена
              </button>
              <button
                onClick={handleConfirmDelete}
                disabled={deleteLoading}
                className="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white text-sm rounded-lg transition"
              >
                {deleteLoading ? 'Удаление...' : 'Удалить'}
              </button>
            </div>
          </div>
        </div>
      )}

      <header className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-white">Приглашение на КВН в ОФТИ</h1>
            <p className="text-gray-400 text-sm">{user?.username} ({user?.email})</p>
          </div>
          <button
            onClick={logout}
            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded-lg text-sm transition"
          >
            Выйти
          </button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <div className="flex gap-3 mb-6">
          <a
            href="/invites/activate"
            className="px-4 py-2 bg-gray-800 hover:bg-gray-700 border border-gray-700 text-gray-300 rounded-lg text-sm transition"
          >
            Активировать инвайт
          </a>
          {user && user.role_id >= 2 && (
            <a
              href="/admin/invites"
              className="px-4 py-2 bg-gray-800 hover:bg-gray-700 border border-gray-700 text-gray-300 rounded-lg text-sm transition"
            >
              Создать инвайт
            </a>
          )}
          {user && user.role_id >= 5 && (
            <a
              href="/admin/users"
              className="px-4 py-2 bg-gray-800 hover:bg-gray-700 border border-gray-700 text-gray-300 rounded-lg text-sm transition"
            >
              Пользователи
            </a>
          )}
        </div>

        {loading ? (
          <div className="text-gray-400">Загрузка...</div>
        ) : error ? (
          <div className="p-4 bg-red-500/20 border border-red-500 rounded-lg text-red-300">
            {error}
          </div>
        ) : (
          <>
            {renderClientList(
              vlessClients,
              'vless',
              'VLESS-клиенты',
              'У вас пока нет VLESS-клиентов.',
            )}
            {renderClientList(
              hysteriaClients,
              'hysteria',
              'Hysteria-клиенты',
              'У вас пока нет Hysteria-клиентов.',
            )}
          </>
        )}
      </main>
    </div>
  );
}