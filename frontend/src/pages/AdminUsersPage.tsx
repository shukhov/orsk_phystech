import { useState, type FormEvent } from 'react';
import { getUserById, setRoleForUser } from '@/api/client';
import { useAuth } from '@/context/AuthContext';
import type { UserPublicOut, UserPrivateOut } from '@/types';

export default function AdminUsersPage() {
  const { user } = useAuth();
  const [searchId, setSearchId] = useState('');
  const [foundUser, setFoundUser] = useState<UserPublicOut | UserPrivateOut | null>(null);
  const [roleId, setRoleId] = useState('');
  const [loading, setLoading] = useState(false);
  const [searchLoading, setSearchLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const handleSearch = async (e: FormEvent) => {
    e.preventDefault();
    if (!searchId) return;
    setError('');
    setFoundUser(null);
    setSuccess('');
    setSearchLoading(true);
    try {
      const data = await getUserById(Number(searchId));
      setFoundUser(data);
    } catch (err: any) {
      setError(err.message || 'Пользователь не найден');
    } finally {
      setSearchLoading(false);
    }
  };

  const handleSetRole = async (e: FormEvent) => {
    e.preventDefault();
    if (!foundUser || !roleId) return;
    setError('');
    setSuccess('');
    setLoading(true);
    try {
      const updated = await setRoleForUser(foundUser.id, Number(roleId));
      setFoundUser(updated);
      setSuccess(`Роль успешно изменена на role_id=${updated.role_id}`);
      setRoleId('');
    } catch (err: any) {
      setError(err.message || 'Ошибка изменения роли');
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
        <h1 className="text-2xl font-bold text-white mb-6">Управление пользователями</h1>

        {error && (
          <div className="mb-4 p-3 bg-red-500/20 border border-red-500 rounded-lg text-red-300 text-sm">
            {error}
          </div>
        )}

        {success && (
          <div className="mb-4 p-3 bg-green-500/20 border border-green-500 rounded-lg text-green-300 text-sm">
            {success}
          </div>
        )}

        {/* Поиск пользователя */}
        <form onSubmit={handleSearch} className="mb-6">
          <label className="block text-gray-300 text-sm mb-1">Найти пользователя по ID</label>
          <div className="flex gap-2">
            <input
              type="number"
              value={searchId}
              onChange={(e) => setSearchId(e.target.value)}
              required
              placeholder="ID пользователя"
              className="flex-1 px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
            />
            <button
              type="submit"
              disabled={searchLoading}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm rounded-lg transition"
            >
              {searchLoading ? '...' : 'Найти'}
            </button>
          </div>
        </form>

        {/* Найденный пользователь */}
        {foundUser && (
          <div className="bg-gray-800 border border-gray-700 rounded-xl p-4 mb-6">
            <h3 className="text-white font-medium mb-2">Пользователь</h3>
            <div className="text-sm text-gray-300 space-y-1">
              <p>ID: {foundUser.id}</p>
              <p>Имя: {foundUser.username}</p>
              <p>Зарегистрирован: {new Date(foundUser.created_at).toLocaleDateString('ru-RU')}</p>
              {'role_id' in foundUser && (
                <p>Текущая роль: role_id = {(foundUser as UserPrivateOut).role_id}</p>
              )}
              {'status' in foundUser && (
                <p>Статус: {(foundUser as UserPrivateOut).status}</p>
              )}
            </div>

            {/* Форма смены роли */}
            <form onSubmit={handleSetRole} className="mt-4 pt-4 border-t border-gray-700">
              <label className="block text-gray-300 text-sm mb-1">Назначить роль</label>
              <div className="flex gap-2">
                <input
                  type="number"
                  value={roleId}
                  onChange={(e) => setRoleId(e.target.value)}
                  required
                  placeholder="role_id"
                  min="1"
                  max="5"
                  className="flex-1 px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                />
                <button
                  type="submit"
                  disabled={loading}
                  className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white text-sm rounded-lg transition"
                >
                  {loading ? '...' : 'Назначить'}
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Справка по ролям */}
        <div className="bg-gray-800 border border-gray-700 rounded-xl p-4">
          <h3 className="text-white font-medium mb-2">Справка по ролям</h3>
          <div className="text-sm text-gray-400 space-y-1">
            <p><span className="text-gray-300">1</span> — базовый пользователь</p>
            <p><span className="text-gray-300">2</span> — создание инвайтов</p>
            <p><span className="text-gray-300">4</span> — просмотр конфига и ролей</p>
            <p><span className="text-gray-300">5</span> — администратор (назначение ролей)</p>
          </div>
        </div>
      </main>
    </div>
  );
}