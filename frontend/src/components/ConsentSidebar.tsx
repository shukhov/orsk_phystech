import { useState } from 'react';

export default function ConsentSidebar() {
  const [open, setOpen] = useState(true);
  const [dismissed, setDismissed] = useState(false);

  if (dismissed) return null;

  return (
    <>
      {/* === Десктоп: боковое окно === */}
      <div className="hidden md:block">
        {/* Кнопка открытия */}
        {!open && (
          <button
            onClick={() => setOpen(true)}
            className="fixed left-0 top-1/2 -translate-y-1/2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium px-2 py-4 rounded-r-lg shadow-lg transition z-40"
            style={{ writingMode: 'vertical-rl', textOrientation: 'mixed' }}
          >
            Персональные данные
          </button>
        )}

        {/* Затемнение */}
        {open && (
          <div
            className="fixed inset-0 bg-black/40 z-40"
            onClick={() => setOpen(false)}
          />
        )}

        {/* Боковая панель */}
        <div
          className={`fixed top-0 left-0 h-full w-80 bg-gray-800 border-r border-gray-700 shadow-2xl z-50 transform transition-transform duration-300 ${
            open ? 'translate-x-0' : '-translate-x-full'
          }`}
        >
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-white">Согласие</h2>
              <button
                onClick={() => setOpen(false)}
                className="text-gray-400 hover:text-white transition"
              >
                ✕
              </button>
            </div>
            <div className="bg-gray-900 rounded-lg p-4 border border-gray-700">
              <p className="text-gray-300 text-sm leading-relaxed">
                Сервис хранит ваши персональные данные: e-mail. Продолжая использование сервиса вы даете согласие на хранение и обработку персональных данных.
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* === Мобильные: баннер внизу === */}
      <div className="md:hidden">
        <div className="fixed bottom-0 left-0 right-0 bg-gray-800 border-t border-gray-700 shadow-2xl z-50">
          <div className="p-4">
            <div className="flex items-start justify-between gap-3">
              <p className="text-gray-300 text-xs leading-relaxed">
                Сервис хранит ваши персональные данные: e-mail. Продолжая использование сервиса вы даете согласие на хранение и обработку персональных данных.
              </p>
              <button
                onClick={() => setDismissed(true)}
                className="text-gray-400 hover:text-white transition shrink-0 text-lg leading-none"
              >
                ✕
              </button>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
