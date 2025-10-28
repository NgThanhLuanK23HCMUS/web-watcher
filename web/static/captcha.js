const captchaContainer = document.createElement("div");
captchaContainer.style.position = "relative";
captchaContainer.style.width = "300px";
captchaContainer.style.height = "150px";
captchaContainer.style.border = "1px solid #ccc";

const puzzleCanvas = document.createElement("canvas");
const sliderCanvas = document.createElement("canvas");

puzzleCanvas.width = 300;
puzzleCanvas.height = 150;
sliderCanvas.width = 50;
sliderCanvas.height = 50;


const targetPositionX = Math.floor(Math.random() * 200) + 50; // Random position between 50-250px
const targetPositionY = 50; // Fixed Y position for slider piece

puzzleCanvas.style.position = "absolute";
sliderCanvas.style.position = "absolute";
sliderCanvas.style.left = "-100px";
sliderCanvas.style.top = targetPositionY + "px";
sliderCanvas.style.cursor = "pointer";

const puzzleCtx = puzzleCanvas.getContext("2d");
const sliderCtx = sliderCanvas.getContext("2d");

const img = new Image();
img.crossOrigin = "anonymous";
img.src = "/static/images/house-and-flower.jpg";

let isDragging = false;
let startX;
let sliderLeft = 0;


const pieceSize = 50;

// Add debug logging and error handling for image
img.onerror = function() {
    console.error('Failed to load image:', img.src);
};

img.onload = function() {
    console.log('Image loaded successfully, size:', img.width, 'x', img.height);
    
    // Clear both canvases first
    puzzleCtx.clearRect(0, 0, puzzleCanvas.width, puzzleCanvas.height);
    sliderCtx.clearRect(0, 0, sliderCanvas.width, sliderCanvas.height);
    
    // Draw main image
    puzzleCtx.drawImage(img, 0, 0, puzzleCanvas.width, puzzleCanvas.height);
    
    try {
        // First capture the image data for the slider piece
        const tempCanvas = document.createElement('canvas');
        tempCanvas.width = pieceSize;
        tempCanvas.height = pieceSize;
        const tempCtx = tempCanvas.getContext('2d');
        
        // Draw the portion we need into the temporary canvas
        tempCtx.drawImage(
            img,
            (targetPositionX / puzzleCanvas.width) * img.width,    // scale x position to image coordinates
            (targetPositionY / puzzleCanvas.height) * img.height,  // scale y position to image coordinates
            (pieceSize / puzzleCanvas.width) * img.width,         // scale width to image coordinates
            (pieceSize / puzzleCanvas.height) * img.height,       // scale height to image coordinates
            0, 0, pieceSize, pieceSize
        );
        
        // Draw the temporary canvas to the slider
        sliderCtx.drawImage(tempCanvas, 0, 0);
        
        // Now create the hole in the main image
        puzzleCtx.fillStyle = "rgba(255, 255, 255, 0.8)";
        puzzleCtx.fillRect(targetPositionX, targetPositionY, pieceSize, pieceSize);
        puzzleCtx.strokeStyle = '#000000';
        puzzleCtx.lineWidth = 2;
        puzzleCtx.strokeRect(targetPositionX, targetPositionY, pieceSize, pieceSize);
        
        // Add a white border to the slider piece
        sliderCtx.strokeStyle = '#ffffff';
        sliderCtx.lineWidth = 2;
        sliderCtx.strokeRect(0, 0, pieceSize, pieceSize);
        
        console.log('Drawing completed successfully');
    } catch (e) {
        console.error('Error drawing slider piece:', e);
    }
}

// Add drag functionality
sliderCanvas.addEventListener('mousedown', function(e) {
    isDragging = true;
    startX = e.clientX - sliderCanvas.offsetLeft;
});

document.addEventListener('mousemove', function(e) {
    if (!isDragging) return;
    
    e.preventDefault();
    const x = e.clientX - startX;
    
    // Constrain slider movement
    if (x < 0) sliderLeft = 0;
    else if (x > 250) sliderLeft = 250;
    else sliderLeft = x;
    
    sliderCanvas.style.left = sliderLeft + 'px';
});

document.addEventListener('mouseup', function() {
    if (!isDragging) return;
    isDragging = false;
    
    // Check if slider is in correct position (within 5px tolerance)
    if (Math.abs(sliderLeft - targetPositionX) < 5) {
        // Add visual feedback first
        captchaContainer.style.border = "2px solid #4CAF50";
        
        fetch('/verify-captcha', {method: 'POST'}) 
        .then(response => {
            if (response.ok) {
                window.location.href = "/login";
            } else {
                console.error('Captcha verfication failed');
            }
        })
        .catch(error => {console.error('Error during captcha verification:', error)});
    } else {
        // Reset slider and show error
        sliderCanvas.style.left = '0px';
        captchaContainer.style.border = "2px solid #ff0000";
        setTimeout(() => {
            captchaContainer.style.border = "1px solid #ccc";
        }, 1000);
    }
});

// Create a wrapper div for the captcha
const captchaWrapper = document.createElement("div");
captchaWrapper.style.margin = "20px 0";
captchaWrapper.style.display = "flex";
captchaWrapper.style.flexDirection = "column";
captchaWrapper.style.alignItems = "center";

// Add a title for the captcha
const captchaTitle = document.createElement("h3");
captchaTitle.textContent = "Complete the Captcha";
captchaTitle.style.marginBottom = "10px";

// Append elements
captchaContainer.appendChild(puzzleCanvas);
captchaContainer.appendChild(sliderCanvas);
captchaWrapper.appendChild(captchaTitle);
captchaWrapper.appendChild(captchaContainer);

// Append to body or a specific container
document.body.appendChild(captchaWrapper);