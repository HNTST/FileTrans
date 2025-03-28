document.addEventListener('DOMContentLoaded', () => {
    // Загрузка списка файлов при старте
    fetchFiles();

    // Обработка формы загрузки
    document.getElementById('uploadForm').addEventListener('submit', async (e) => {
        e.preventDefault();

        const fileInput = document.getElementById('fileInput');
        const file = fileInput.files[0];
        const status = document.getElementById('uploadStatus');

        if (!file) {
            status.textContent = 'Please select a file!';
            return;
        }

        const formData = new FormData();
        formData.append('file', file);

        try {
            status.textContent = 'Uploading...';
            const response = await fetch('/upload', {
                method: 'POST',
                body: formData,
            });

            if (!response.ok) {
                throw new Error('Upload failed');
            }

            const data = await response.json();
            status.textContent = `File "${data.filename}" uploaded successfully!`;
            fileInput.value = ''; // Очистка поля
            fetchFiles(); // Обновление списка файлов
        } catch (error) {
            status.textContent = `Error: ${error.message}`;
        }
    });
});

// Получение и отображение списка файлов
async function fetchFiles() {
    try {
        const response = await fetch('/files');
        const fileList = document.getElementById('fileList');

        if (!response.ok) {
            throw new Error('Failed to fetch files');
        }

        const data = await response.json();

        // Если вернулся объект с сообщением
        if (data.message) {
            fileList.innerHTML = '<li>No files available</li>';
            return;
        }

        // Отображение списка файлов
        fileList.innerHTML = '';
        data.forEach(file => {
            const li = document.createElement('li');
            const a = document.createElement('a');
            a.href = `/download/${file.id}`;
            a.textContent = `${file.filename} (ID: ${file.id})`;
            a.download = file.filename; // Указываем имя файла для скачивания
            li.appendChild(a);
            fileList.appendChild(li);
        });
    } catch (error) {
        console.error('Error fetching files:', error);
        document.getElementById('fileList').innerHTML = '<li>Error loading files</li>';
    }
}