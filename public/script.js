document.addEventListener('DOMContentLoaded', () => {
    fetchFiles();

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
            fileInput.value = '';
            fetchFiles();
        } catch (error) {
            status.textContent = `Error: ${error.message}`;
        }
    });
});

async function fetchFiles() {
    try {
        const response = await fetch('/files');
        const fileList = document.getElementById('fileList');

        if (!response.ok) {
            throw new Error('Failed to fetch files');
        }

        const data = await response.json();
        console.log('Данные от сервера:', data); // Для отладки

        // Проверяем, является ли data массивом
        let files = Array.isArray(data) ? data : data.files || data.data || [];

        if (!files.length || data.message === "No files available") {
            fileList.innerHTML = '<li>No files available</li>';
            return;
        }

        fileList.innerHTML = '';
        files.forEach(file => {
            const li = document.createElement('li');
            const a = document.createElement('a');
            a.href = `/download/${file.id}`;
            a.textContent = `${file.filename} (ID: ${file.id})`;
            a.download = file.filename;
            li.appendChild(a);
            fileList.appendChild(li);
        });
    } catch (error) {
        console.error('Error fetching files:', error);
        fileList.innerHTML = '<li>Error loading files</li>';
    }
}