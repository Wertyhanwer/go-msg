const API_BASE = '/api';

// Состояние приложения
let currentUser = null;
let selectedFriend = null;
let messageUpdateInterval = null;
let websocket = null;

// Показ статуса
function showStatus(message, isError = false) {
    const status = document.getElementById('status');
    status.textContent = message;
    status.className = isError ? 'status error' : 'status success';
    setTimeout(() => {
        status.textContent = '';
        status.className = 'status';
    }, 3000);
}

// Переключение вкладок авторизации
function switchTab(tab) {
    const loginTab = document.getElementById('loginTab');
    const registerTab = document.getElementById('registerTab');
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');

    if (tab === 'login') {
        loginTab.classList.add('active');
        registerTab.classList.remove('active');
        loginForm.classList.remove('hidden');
        registerForm.classList.add('hidden');
    } else {
        registerTab.classList.add('active');
        loginTab.classList.remove('active');
        registerForm.classList.remove('hidden');
        loginForm.classList.add('hidden');
    }
}

// Регистрация
async function register() {
    const firstName = document.getElementById('registerFirstName').value.trim();
    const lastName = document.getElementById('registerLastName').value.trim();
    const email = document.getElementById('registerEmail').value.trim();
    const password = document.getElementById('registerPassword').value;

    if (!firstName || !lastName || !email || !password) {
        showStatus('Заполните все поля', true);
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ first_name: firstName, last_name: lastName, email, password }),
            credentials: 'include'
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        showStatus('Регистрация успешна! Теперь войдите в систему');
        switchTab('login');
        
        // Очищаем форму
        document.getElementById('registerFirstName').value = '';
        document.getElementById('registerLastName').value = '';
        document.getElementById('registerEmail').value = '';
        document.getElementById('registerPassword').value = '';
    } catch (error) {
        showStatus(`Ошибка регистрации: ${error.message}`, true);
    }
}

// Вход
async function login() {
    const email = document.getElementById('loginEmail').value.trim();
    const password = document.getElementById('loginPassword').value;

    if (!email || !password) {
        showStatus('Введите email и пароль', true);
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password }),
            credentials: 'include'
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        await loadCurrentUser();
        showChatScreen();
        showStatus('Вход выполнен успешно!');
    } catch (error) {
        showStatus(`Ошибка входа: ${error.message}`, true);
    }
}

// Выход
async function logout() {
    try {
        await fetch(`${API_BASE}/auth/logout`, {
            method: 'POST',
            credentials: 'include'
        });
    } catch (error) {
        console.error('Ошибка выхода:', error);
    }

    currentUser = null;
    selectedFriend = null;
    if (messageUpdateInterval) {
        clearInterval(messageUpdateInterval);
        messageUpdateInterval = null;
    }
    if (websocket) {
        websocket.close();
        websocket = null;
    }
    showAuthScreen();
    showStatus('Вы вышли из системы');
}

// Загрузка текущего пользователя
async function loadCurrentUser() {
    console.log('loadCurrentUser called');
    try {
        const response = await fetch(`${API_BASE}/auth/user`, {
            credentials: 'include'
        });

        console.log('loadCurrentUser response status:', response.status);

        if (!response.ok) {
            throw new Error('Не авторизован');
        }

        currentUser = await response.json();
        console.log('Current user loaded:', currentUser);
        document.getElementById('currentUserName').textContent = 
            `${currentUser.first_name} ${currentUser.last_name}`;
        
        return currentUser;
    } catch (error) {
        console.error('Ошибка загрузки пользователя:', error);
        throw error;
    }
}

// Показ экрана авторизации
function showAuthScreen() {
    document.getElementById('authScreen').style.display = 'flex';
    document.getElementById('chatScreen').classList.remove('active');
}

// Показ экрана чата
function showChatScreen() {
    document.getElementById('authScreen').style.display = 'none';
    document.getElementById('chatScreen').classList.add('active');
    
    loadFriends();
    loadFriendRequests();
    connectWebSocket();
    
    // Запускаем автообновление для друзей (сообщения теперь обновляются через WebSocket)
    if (messageUpdateInterval) {
        clearInterval(messageUpdateInterval);
    }
    messageUpdateInterval = setInterval(() => {
        loadFriendRequests();
    }, 30000); // Реже обновляем заявки
}

// Поиск пользователей
let searchTimeout;
async function searchUsers() {
    const query = document.getElementById('userSearch').value.trim();
    const searchResults = document.getElementById('searchResults');

    if (searchTimeout) {
        clearTimeout(searchTimeout);
    }

    if (!query) {
        searchResults.innerHTML = '';
        return;
    }

    searchTimeout = setTimeout(async () => {
        try {
            const response = await fetch(`${API_BASE}/friends/search?q=${encodeURIComponent(query)}`, {
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Ошибка поиска');
            }

            const users = await response.json();
            displaySearchResults(users);
        } catch (error) {
            showStatus(`Ошибка поиска: ${error.message}`, true);
        }
    }, 300);
}

// Отображение результатов поиска
function displaySearchResults(users) {
    const searchResults = document.getElementById('searchResults');
    
    if (!users || users.length === 0) {
        searchResults.innerHTML = '<div style="padding: 10px; color: #6c757d; font-size: 14px;">Пользователи не найдены</div>';
        return;
    }

    searchResults.innerHTML = users.map(user => `
        <div class="search-result-item">
            <div>
                <div style="font-weight: 600;">${user.first_name} ${user.last_name}</div>
                <div style="font-size: 12px; color: #6c757d;">${user.email}</div>
            </div>
            <button class="add-friend-btn" onclick="sendFriendRequest(${user.id})">
                Добавить
            </button>
        </div>
    `).join('');
}

// Отправка заявки в друзья
async function sendFriendRequest(friendId) {
    try {
        const response = await fetch(`${API_BASE}/friends/request`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ friend_id: friendId }),
            credentials: 'include'
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        showStatus('Заявка в друзья отправлена!');
        document.getElementById('userSearch').value = '';
        document.getElementById('searchResults').innerHTML = '';
    } catch (error) {
        showStatus(`Ошибка: ${error.message}`, true);
    }
}

// Загрузка списка друзей
async function loadFriends() {
    try {
        const response = await fetch(`${API_BASE}/friends`, {
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Ошибка загрузки друзей');
        }

        const friends = await response.json();
        displayFriends(friends);
    } catch (error) {
        console.error('Ошибка загрузки друзей:', error);
    }
}

// Отображение списка друзей
function displayFriends(friends) {
    const friendsList = document.getElementById('friendsList');
    
    if (!friends || friends.length === 0) {
        friendsList.innerHTML = '<div style="color: #6c757d; font-size: 14px;">Друзей пока нет</div>';
        return;
    }

    friendsList.innerHTML = friends.map(friend => `
        <div class="friend-item" onclick="selectFriend(${friend.friend_id}, '${friend.first_name}', '${friend.last_name}', '${friend.email}')">
            <div class="friend-name">${friend.first_name} ${friend.last_name}</div>
            <div class="friend-email">${friend.email}</div>
        </div>
    `).join('');
}

// Выбор друга для чата
function selectFriend(friendId, firstName, lastName, email) {
    console.log('selectFriend called with:', {friendId, firstName, lastName, email});
    selectedFriend = { id: friendId, firstName, lastName, email };
    console.log('selectedFriend set to:', selectedFriend);
    
    // Обновляем UI
    document.querySelectorAll('.friend-item').forEach(item => {
        item.classList.remove('active');
    });
    event.currentTarget.classList.add('active');
    
    document.querySelector('.chat-title').textContent = `${firstName} ${lastName}`;
    document.getElementById('messageForm').classList.remove('hidden');
    
    // Загружаем сообщения
    console.log('Loading messages for friend:', friendId);
    loadMessages(friendId);
}

// Загрузка сообщений между пользователями
async function loadMessages(friendId) {
    console.log('loadMessages called with friendId:', friendId);
    try {
        const response = await fetch(`${API_BASE}/messages/between?recipient_id=${friendId}&limit=50&offset=0`, {
            credentials: 'include'
        });

        console.log('loadMessages response status:', response.status);
        
        if (!response.ok) {
            const errorText = await response.text();
            console.log('loadMessages error:', errorText);
            throw new Error('Ошибка загрузки сообщений');
        }

        const messages = await response.json();
        console.log('Loaded messages:', messages);
        displayMessages(messages);
    } catch (error) {
        console.error('Ошибка загрузки сообщений:', error);
        document.getElementById('messagesContainer').innerHTML = 
            '<div class="no-chat-selected">Ошибка загрузки сообщений</div>';
    }
}

// Отображение сообщений
function displayMessages(messages) {
    console.log('displayMessages called with:', messages);
    const container = document.getElementById('messagesContainer');
    
    if (!messages || messages.length === 0) {
        console.log('No messages to display');
        container.innerHTML = '<div class="no-chat-selected">Сообщений пока нет</div>';
        return;
    }

    console.log('Displaying', messages.length, 'messages');
    console.log('Current user ID:', currentUser ? currentUser.id : 'null');

    container.innerHTML = messages.map(message => {
        const isOwn = message.user_id === currentUser.id;
        const time = new Date(message.created_at).toLocaleTimeString('ru-RU', { 
            hour: '2-digit', 
            minute: '2-digit' 
        });
        
        console.log(`Message ${message.id}: isOwn=${isOwn}, user_id=${message.user_id}, current_user_id=${currentUser.id}`);
        
        return `
            <div class="message ${isOwn ? 'own' : ''}">
                <div class="message-content">${escapeHtml(message.text)}</div>
                <div class="message-time">${time}</div>
            </div>
        `;
    }).join('');
    
    // Прокручиваем к последнему сообщению
    container.scrollTop = container.scrollHeight;
    console.log('Messages displayed successfully');
}

// Отправка сообщения
async function sendMessage() {
    console.log('sendMessage called');
    console.log('selectedFriend:', selectedFriend);
    console.log('currentUser:', currentUser);
    
    if (!selectedFriend) {
        showStatus('Выберите собеседника', true);
        console.log('No selected friend');
        return;
    }

    const textElement = document.getElementById('messageText');
    const text = textElement.value.trim();
    
    console.log('Message text:', text);
    
    if (!text) {
        showStatus('Введите текст сообщения', true);
        console.log('Empty message text');
        return;
    }

    try {
        console.log('Sending message to API...');
        const response = await fetch(`${API_BASE}/messages/create`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                recipient_id: selectedFriend.id, 
                text: text 
            }),
            credentials: 'include'
        });

        console.log('API response status:', response.status);
        
        if (!response.ok) {
            const error = await response.text();
            console.log('API error:', error);
            throw new Error(error);
        }

        const result = await response.json();
        console.log('Message created:', result);
        
        textElement.value = '';
        showStatus('Сообщение отправлено!');
        
        // Принудительно обновляем сообщения
        setTimeout(() => {
            loadMessages(selectedFriend.id);
        }, 100);
        
    } catch (error) {
        console.error('Send message error:', error);
        showStatus(`Ошибка отправки: ${error.message}`, true);
    }
}

// Загрузка заявок в друзья
async function loadFriendRequests() {
    try {
        const response = await fetch(`${API_BASE}/friends/pending`, {
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Ошибка загрузки заявок');
        }

        const requests = await response.json();
        displayFriendRequests(requests);
    } catch (error) {
        console.error('Ошибка загрузки заявок:', error);
    }
}

// Отображение заявок в друзья
function displayFriendRequests(requests) {
    const requestsContainer = document.getElementById('friendRequests');
    
    if (!requests || requests.length === 0) {
        requestsContainer.innerHTML = '<div style="color: #6c757d; font-size: 14px;">Новых заявок нет</div>';
        return;
    }

    requestsContainer.innerHTML = requests.map(request => `
        <div class="request-item">
            <div class="friend-name">${request.first_name} ${request.last_name}</div>
            <div class="friend-email">${request.email}</div>
            <div class="request-actions">
                <button class="accept-btn" onclick="acceptFriendRequest(${request.user_id})">
                    Принять
                </button>
                <button class="reject-btn" onclick="rejectFriendRequest(${request.user_id})">
                    Отклонить
                </button>
            </div>
        </div>
    `).join('');
}

// Принятие заявки в друзья
async function acceptFriendRequest(userId) {
    try {
        const response = await fetch(`${API_BASE}/friends/accept`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ friend_id: userId }),
            credentials: 'include'
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        showStatus('Заявка принята!');
        loadFriends();
        loadFriendRequests();
    } catch (error) {
        showStatus(`Ошибка: ${error.message}`, true);
    }
}

// Отклонение заявки в друзья
async function rejectFriendRequest(userId) {
    try {
        const response = await fetch(`${API_BASE}/friends/reject`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ friend_id: userId }),
            credentials: 'include'
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        showStatus('Заявка отклонена');
        loadFriendRequests();
    } catch (error) {
        showStatus(`Ошибка: ${error.message}`, true);
    }
}

// Экранирование HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Обработка нажатия Enter в поле сообщения
document.addEventListener('DOMContentLoaded', function() {
    const messageInput = document.getElementById('messageText');
    if (messageInput) {
        messageInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendMessage();
            }
        });
    }

    // Проверяем авторизацию при загрузке страницы
    checkAuth();
});

// WebSocket подключение
function connectWebSocket() {
    if (websocket) {
        websocket.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/ws`;
    
    websocket = new WebSocket(wsUrl);
    
    websocket.onopen = function() {
        console.log('WebSocket connected');
    };
    
    websocket.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            handleWebSocketMessage(data);
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    };
    
    websocket.onclose = function() {
        console.log('WebSocket disconnected');
        // Переподключаемся через 3 секунды
        setTimeout(() => {
            if (currentUser) {
                connectWebSocket();
            }
        }, 3000);
    };
    
    websocket.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
}

// Обработка WebSocket сообщений
function handleWebSocketMessage(data) {
    switch (data.type) {
        case 'new_message':
            // Если сообщение от текущего собеседника, обновляем чат
            if (selectedFriend && 
                (data.message.user_id === selectedFriend.id || 
                 data.message.recipient_id === selectedFriend.id)) {
                loadMessages(selectedFriend.id);
            }
            
            // Показываем уведомление если сообщение не от нас
            if (data.message.user_id !== currentUser.id) {
                showStatus(`Новое сообщение от ${data.message.sender_name || 'пользователя'}`);
                
                // Обновляем список друзей чтобы показать новое сообщение
                loadFriends();
            }
            break;
            
        default:
            console.log('Unknown WebSocket message type:', data.type);
    }
}

// Проверка авторизации
async function checkAuth() {
    console.log('checkAuth called');
    try {
        await loadCurrentUser();
        console.log('User authenticated, showing chat screen');
        showChatScreen();
    } catch (error) {
        console.log('User not authenticated, showing auth screen');
        showAuthScreen();
    }
}


