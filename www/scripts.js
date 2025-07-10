// Используем относительные URL для работы через прокси
const API_BASE = '/api';

function showStatus(message, isError = false) {
    const status = document.getElementById('status');
    status.textContent = message;
    status.className = isError ? 'status error' : 'status success';
    setTimeout(() => {
        status.textContent = '';
        status.className = 'status';
    }, 3000);
}

async function fetchApi() {
    try {
        const response = await fetch(`${API_BASE}/`);
        if (!response.ok) {
            throw new Error(`Ошибка HTTP: ${response.status}`);
        }
        const data = await response.json();
        showStatus(`API статус: ${data.message}`);
        return data;
    } catch (error) {
        showStatus(`Ошибка API: ${error.message}`, true);
        throw error;
    }
}

async function createMessage() {
    const textElement = document.getElementById('messageText');
    const text = textElement.value.trim();
    
    if (!text) {
        showStatus('Введите текст сообщения', true);
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/messages/create`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ text: text })
        });

        if (!response.ok) {
            throw new Error(`Ошибка HTTP: ${response.status}`);
        }

        const message = await response.json();
        showStatus('Сообщение создано успешно!');
        textElement.value = '';
        loadMessages();
    } catch (error) {
        showStatus(`Ошибка создания: ${error.message}`, true);
    }
}

async function loadMessages() {
    try {
        const response = await fetch(`${API_BASE}/messages?limit=20&offset=0`);
        if (!response.ok) {
            throw new Error(`Ошибка HTTP: ${response.status}`);
        }

        const messages = await response.json();
        displayMessages(messages);
        showStatus(`Загружено ${messages.length} сообщений`);
    } catch (error) {
        document.getElementById('output').innerHTML = 
            `<div class="error">Ошибка загрузки: ${error.message}</div>`;
        showStatus(`Ошибка загрузки: ${error.message}`, true);
    }
}

function displayMessages(messages) {
    const output = document.getElementById('output');
    
    if (!messages || messages.length === 0) {
        output.innerHTML = '<div class="no-messages">Сообщений пока нет</div>';
        return;
    }

    const messagesList = messages.map(message => `
        <div class="message" data-id="${message.id}">
            <div class="message-content">${escapeHtml(message.text)}</div>
            <div class="message-meta">
                ID: ${message.id} | 
                Создано: ${new Date(message.created_at).toLocaleString('ru-RU')}
                <button onclick="deleteMessage(${message.id})" class="delete-btn">Удалить</button>
            </div>
        </div>
    `).join('');

    output.innerHTML = messagesList;
}

async function deleteMessage(id) {
    if (!confirm('Удалить сообщение?')) return;

    try {
        const response = await fetch(`${API_BASE}/messages/delete?id=${id}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            throw new Error(`Ошибка HTTP: ${response.status}`);
        }

        showStatus('Сообщение удалено');
        loadMessages();
    } catch (error) {
        showStatus(`Ошибка удаления: ${error.message}`, true);
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Загружаем данные при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    fetchApi();
    loadMessages();
});


