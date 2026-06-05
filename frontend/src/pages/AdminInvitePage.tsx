import { useState, type FormEvent } from 'react';
import { newInvite } from '@/api/client';
import { useAuth } from '@/context/AuthContext';
import type { InviteOut } from '@/types';

export default function AdminInvitePage() {
  const { user } = useAuth();
  const [inviteWord, setInviteWord] = useState('');
  const [vpnType, setVpnType] = useState('vless');
  const [expiresAt, setExpiresAt] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<InviteOut | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setResult(null);
    setLoading(true);
    try {
      const data = await newInvite({
        invite_word: inviteWord,
        vpn_type: vpnType,
        expires_at: expiresAt ? new Date(expiresAt).toISOString() : undefined,
      });
      setResult(data);
      setInviteWord('');
    } catch (err: any) {
      setError(err.message || 'Ошибка создания инвайта');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900">
      <header className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
          <a href="/" className="text-xl font-bold text-white hover:text-gray-300 transition">
            Приглашение на КВН в ОФТИ
          </a>
          <span className="text-gray-400 text-sm">{user?.username}</span>
        </div>
      </header>

      <main className="max-w-lg mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold text-white mb-6">Создание инвайта</h1>

        {error && (
          <div className="mb-4 p-3 bg-red-500/20 border border-red-500 rounded-lg text-red-300 text-sm">
            {error}
          </div>
        )}

        {result && (
          <div className="mb-6 p-4 bg-green-500/20 border border-green-500 rounded-lg text-green-300">
            <p className="font-medium mb-2">Инвайт создан!</p>
            <div className="text-sm space-y-1">
              <p>ID: {result.id}</p>
              <p>Тип VPN: {result.vpn_type}</p>
              <p>Статус: {result.status}</p>
              <p>Истекает: {new Date(result.expires_at).toLocaleDateString('ru-RU')}</p>
            </div>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-gray-300 text-sm mb-1">Инвайт-код</label>
            <input
              type="text"
              value={inviteWord}
              onChange={(e) => setInviteWord(e.target.value)}
              required
              placeholder="Уникальное слово для инвайта"
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
            />
            <p className="text-gray-500 text-xs mt-1">
              Этот код нужно передать пользователю для активации
            </p>
          </div>

          <div>
            <label className="block text-gray-300 text-sm mb-1">Тип VPN</label>
            <select
              value={vpnType}
              onChange={(e) => setVpnType(e.target.value)}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500"
            >
              <option value="vless">VLESS</option>
            </select>
          </div>

          <div>
            <label className="block text-gray-300 text-sm mb-1">Дата истечения (необязательно)</label>
            <input
              type="datetime-local"
              value={expiresAt}
              onChange={(e) => setExpiresAt(e.target.value)}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500"
            />
            <p className="text-gray-500 text-xs mt-1">
              Если не указано, инвайт будет действителен 1 месяц
            </p>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-medium rounded-lg transition"
          >
            {loading ? 'Создание...' : 'Создать инвайт'}
          </button>
        </form>
      </main>
    </div>
  );
}