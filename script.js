let currentImagePath = '';
let currentSelectedImage = null; // Track the current selected image element

function loadImages() {
    fetch('/images')
        .then(response => response.json())
        .then(data => {
            const images = data.images;
            const total = data.total;
            const totalInfo = document.getElementById('totalInfo');
            const totalSizeInfo = document.getElementById('totalSize');
            const imagesContainer = document.getElementById('imagesContainer');
            imagesContainer.innerHTML = '';

            totalInfo.innerText = `Total number of files: ${total}`;

            let totalSize = 0;
            images.forEach(image => {
                const imageWrapper = document.createElement('div');
                imageWrapper.classList.add('imageWrapper');

                const skeleton = document.createElement('div');
                skeleton.classList.add('skeleton');

                const img = document.createElement('img');
                img.dataset.src = `/images/${image.name}`;
                img.alt = 'Discord Cache Image';
                img.addEventListener('load', () => {
                    img.classList.add('lazy-loaded');
                    imageWrapper.removeChild(skeleton);
                });
                img.addEventListener('click', () => {
                    openModal(img.dataset.src, image.name);
                    highlightImage(imageWrapper);
                });
                img.addEventListener('contextmenu', (e) => {
                    showContextMenu(e, image.name);
                    highlightImage(imageWrapper);
                });

                const info = document.createElement('div');
                info.classList.add('info');
                info.innerText = `Name: ${image.name}\nModified: ${moment(image.timestamp).fromNow()}\nSize: ${image.size}`;

                const sizeInMB = parseFloat(image.size.split(' ')[0]);
                totalSize += sizeInMB;

                imageWrapper.appendChild(skeleton);
                imageWrapper.appendChild(img);
                imageWrapper.appendChild(info);
                imagesContainer.appendChild(imageWrapper);
            });

            if (totalSize >= 1024) {
                totalSizeInfo.innerText = `Total size: ${(totalSize / 1024).toFixed(2)} GB`;
            } else {
                totalSizeInfo.innerText = `Total size: ${totalSize.toFixed(2)} MB`;
            }

            lazyLoadImages();
        });
}

function highlightImage(imageWrapper) {
    if (currentSelectedImage) {
        currentSelectedImage.classList.remove('selected');
    }
    imageWrapper.classList.add('selected');
    currentSelectedImage = imageWrapper;
}

function openModal(src, name) {
    const modal = document.getElementById('imageModal');
    const modalImg = document.getElementById('modalImage');
    const captionText = document.getElementById('imageCaption');

    modal.style.display = "block";
    modalImg.src = src;
    captionText.innerText = name;
}

function closeModal() {
    const modal = document.getElementById('imageModal');
    modal.style.display = "none";
}

function lazyLoadImages() {
    const lazyImages = document.querySelectorAll('img[data-src]');
    const config = {
        rootMargin: '0px 0px 50px 0px',
        threshold: 0.01
    };

    const imageObserver = new IntersectionObserver((entries, observer) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const img = entry.target;
                img.src = img.dataset.src;
                observer.unobserve(img);
            }
        });
    }, config);

    lazyImages.forEach(image => {
        imageObserver.observe(image);
    });
}

function deleteAllCache() {
    fetch('/delete-cache', { method: 'DELETE' })
        .then(response => {
            if (response.ok) {
                loadImages();
                alert('All cache has been deleted.');
            } else {
                alert('Failed to delete cache.');
            }
        });
}

function showContextMenu(event, imagePath) {
    event.preventDefault();
    const contextMenu = document.getElementById('contextMenu');
    currentImagePath = imagePath;

    contextMenu.style.top = `${event.clientY}px`;
    contextMenu.style.left = `${event.clientX}px`;
    contextMenu.style.display = 'block';

    document.addEventListener('click', hideContextMenu);
}

function hideContextMenu() {
    const contextMenu = document.getElementById('contextMenu');
    contextMenu.style.display = 'none';
    document.removeEventListener('click', hideContextMenu);
}

document.getElementById('openInFileExplorer').addEventListener('click', () => {
    fetch(`/open-file?path=${encodeURIComponent(currentImagePath)}`, { method: 'GET' })
        .then(response => {
            if (!response.ok) {
                alert('Failed to open file in explorer.');
            }
        });
    hideContextMenu();
});

document.getElementById('deleteImage').addEventListener('click', () => {
    fetch(`/delete-image?path=${encodeURIComponent(currentImagePath)}`, { method: 'DELETE' })
        .then(response => {
            if (response.ok) {
                loadImages();
                alert(`Deleted ${currentImagePath}`);
            } else {
                alert('Failed to delete image.');
            }
        });
    hideContextMenu();
});

window.onload = () => {
    loadImages();

    const closeModalBtn = document.getElementById('closeModal');
    closeModalBtn.addEventListener('click', closeModal);

    const deleteCacheButton = document.getElementById('deleteCacheButton');
    deleteCacheButton.addEventListener('click', deleteAllCache);
};