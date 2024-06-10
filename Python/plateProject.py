import cv2
import pytesseract
import os
import numpy as np

#Paths
image_folder_path = 'Liscence Plates'
output_file_path = 'plateNumbers.txt'

def preprocess_image(image_path):
    img = cv2.imread(image_path)

    #Histogram Equalization
    img = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    img = cv2.equalizeHist(img)

    #Gaussian
    gaussian_blur = cv2.GaussianBlur(img, (0, 0), 3)
    unsharp_image = cv2.addWeighted(img, 1.5, gaussian_blur, -0.5, 0)

    #Thresholding
    # _, binary_image = cv2.threshold(unsharp_image, 120, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
    #return binary_image
    return unsharp_image

#Optional (personal challenge)
#Experimental function attemption to extract text from image,
def extract_text_from_image(img):
    img = cv2.resize(img, None, fx=2, fy=2, interpolation=cv2.INTER_CUBIC)
    kernel = np.ones((1, 1), np.uint8)
    img = cv2.dilate(img, kernel, iterations=1)
    custom_config = r'--oem 3 --psm 6 -c tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    text = pytesseract.image_to_string(img, config=custom_config)
    return text.strip()

def edgeDetection(image_path):
    img = cv2.imread(image_path, cv2.IMREAD_COLOR)
    #grayscale
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    bilateral = cv2.bilateralFilter(gray, 9, 75, 75)
    gaussian = cv2.GaussianBlur(bilateral, (9, 9), 10)
    unsharp_image = cv2.addWeighted(bilateral, 1.5, gaussian, -0.5, 0)
    #Canny edges
    edges = cv2.Canny(bilateral, 25, 50)  
    
    return edges

def process_images(folder_path, output_path):
    with open(output_path, 'w') as output_file:
        for i in range(1, 17):
            image_path = os.path.join(folder_path, f"{i}.jpg")
            original_img = cv2.imread(image_path)
            img = preprocess_image(image_path)
            text = extract_text_from_image(img)
            edges = edgeDetection(image_path) 
            
            # Display the images
            cv2.imshow(f'Original Image {i}', original_img)
            cv2.imshow(f'Processed Image {i}', img)
            cv2.imshow(f'Edges {i}', edges)
            cv2.waitKey(0) 
            cv2.destroyAllWindows()  
            

process_images(image_folder_path, output_file_path)
