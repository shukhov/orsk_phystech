import { useState, useEffect } from 'react';
import QRCode from 'qrcode';
import { useAuth } from '@/context/AuthContext';
import { getClientsByUserId, getXrayLink } from '@/api/client';
import type { ClientPublicOut } from '@/types';

export default function DashboardPage() {
  const { user, logout } = useAuth();
  const [clients, setClients] = useState<ClientPublicOut[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activeClientId, setActiveClientId] = useState<number | null>(null);
  const [link, setLink] = useState('');
  const [linkLoading, setLinkLoading] = useState(false);
  const [linkError, setLinkError] = useState('');
  const [copied, setCopied] = useState(false);
  const [qrDataUrl, setQrDataUrl] = useState('');

  useEffect(() => {
    if (!user) return;
    setLoading(true);
    getClientsByUserId(user.id)
      .then(setClients)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [user]);

  const handleGetLink = async (clientId: number) => {
    if (activeClientId === clientId && link) {
      setActiveClientId(null);
      setLink('');
      setQrDataUrl('');
      return;
    }
    setActiveClientId(clientId);
    setLink('');
    setLinkError('');
    setCopied(false);
    setQrDataUrl('');
    setLinkLoading(true);
    try {
      const result = await getXrayLink(clientId);
      setLink(result.connection_link);
      if (result.connection_link.startsWith('vless://')) {
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

  return (
    <div className="min-h-screen bg-gray-900">
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

        <h2 className="text-lg font-semibold text-white mb-4">Ваши VPN-клиенты</h2>

        {loading ? (
          <div className="text-gray-400">Загрузка...</div>
        ) : error ? (
          <div className="p-4 bg-red-500/20 border border-red-500 rounded-lg text-red-300">
            {error}
          </div>
        ) : clients.length === 0 ? (
          <div className="p-6 bg-gray-800 rounded-xl border border-gray-700 text-center">
            <p className="text-gray-400">У вас пока нет VPN-клиентов.</p>
            <a href="/invites/activate" className="text-blue-400 hover:underline text-sm mt-2 inline-block">
              Активируйте инвайт, чтобы создать клиента
            </a>
          </div>
        ) : (
          <div className="space-y-3">
            {clients.map((client) => (
              <div
                key={client.id}
                className="bg-gray-800 border border-gray-700 rounded-xl p-4"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <span className="text-white font-medium">{client.alias}</span>
                    <span className={`ml-3 text-sm ${statusColor(client.status)}`}>
                      {statusLabel(client.status)}
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-gray-500 text-sm">с {formatDate(client.created_at)}</span>
                    <button
                      onClick={() => handleGetLink(client.id)}
                      className="px-4 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded-lg transition"
                    >
                      {activeClientId === client.id && linkLoading ? 'Загрузка...' : 'Получить ссылку'}
                    </button>
                  </div>
                </div>

                {activeClientId === client.id && (
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
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
