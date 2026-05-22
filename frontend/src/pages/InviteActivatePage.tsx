import { useState, type FormEvent } from 'react';
import { activateInvite } from '@/api/client';
import { useAuth } from '@/context/AuthContext';
import type { InviteCheckOut } from '@/types';

export default function InviteActivatePage() {
  const { user } = useAuth();
  const [inviteWord, setInviteWord] = useState('');
  const [alias, setAlias] = useState('');
  const [targetUserId, setTargetUserId] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<InviteCheckOut | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setResult(null);
    setLoading(true);
    try {
      const payload: any = { invite_word: inviteWord, alias };
      if (targetUserId) {
        payload.user_id = Number(targetUserId);
      }
      const data = await activateInvite(payload);
      setResult(data);
    } catch (err: any) {
      setError(err.message || 'Ошибка активации');
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
        <h1 className="text-2xl font-bold text-white mb-6">Активация инвайта</h1>

        {error && (
          <div className="mb-4 p-3 bg-red-500/20 border border-red-500 rounded-lg text-red-300 text-sm">
            {error}
          </div>
        )}

        {result && (
          <div className="mb-6 p-4 bg-green-500/20 border border-green-500 rounded-lg text-green-300">
            <p className="font-medium mb-1">Клиент успешно создан!</p>
            <p className="text-sm">Алиас: <span className="font-mono">{result.alias}</span></p>
            <p className="text-sm">Тип VPN: <span className="font-mono">{result.vpn_type}</span></p>
            <a href="/" className="text-blue-400 hover:underline text-sm mt-2 inline-block">
              Вернуться на главную →
            </a>
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
              placeholder="Введите инвайт-код"
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
            />
          </div>

          <div>
            <label className="block text-gray-300 text-sm mb-1">Алиас (название клиента)</label>
            <input
              type="text"
              value={alias}
              onChange={(e) => setAlias(e.target.value)}
              required
              placeholder="Например: мой-iphone"
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
            />
          </div>

          <div>
            <label className="block text-gray-300 text-sm mb-1">ID пользователя (необязательно)</label>
            <input
              type="number"
              value={targetUserId}
              onChange={(e) => setTargetUserId(e.target.value)}
              placeholder="Оставьте пустым для себя"
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
            />
            <p className="text-gray-500 text-xs mt-1">Если оставить пустым, инвайт активируется для вас</p>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-medium rounded-lg transition"
          >
            {loading ? 'Активация...' : 'Активировать'}
          </button>
        </form>
      </main>
    </div>
  );
}
