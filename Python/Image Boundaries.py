import cv2
import numpy as np
from matplotlib import pyplot as plt


def boundaryParticles(img):
    _, thresh = cv2.threshold(img, 127, 255, cv2.THRESH_BINARY)
    contours, _ = cv2.findContours(thresh, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    height, width = img.shape
    boundary_touching = np.zeros_like(img)

    for contour in contours:
        if cv2.contourArea(contour) > 0:
            x, y, w, h = cv2.boundingRect(contour)
            if x == 0 or y == 0 or (x + w) == width or (y + h) == height:
                cv2.drawContours(boundary_touching, [contour], -1, 255, thickness=cv2.FILLED)

    return boundary_touching



def prob2():
    img = cv2.imread("DIP3E_Original_Images_CH09/FigP0936(bubbles_on_black_background).tif", 0)
    boundaryImg = boundaryParticles(img)
    cv2.imshow("Boundary Touching Particles", boundaryImg)
    # cv2.imshow("Overlapping Particles", overlapImg)
    cv2.waitKey(0)
    cv2.destroyAllWindows()


def prob3():
    #Black image  512×512 white square of size 256×256 at its center
    image_size = 512
    square_size = 256
    image = np.zeros((image_size, image_size), dtype=np.uint8)
    start = (image_size - square_size) // 2
    end = start + square_size
    image[start:end, start:end] = 255

    #gradient using the Sobel 
    sobelx = cv2.Sobel(image, cv2.CV_64F, 1, 0, ksize=3)
    sobely = cv2.Sobel(image, cv2.CV_64F, 0, 1, ksize=3)
    gradient_magnitude = np.sqrt(sobelx**2 + sobely**2)

    # Display the Image
    plt.figure(figsize=(6, 6))
    plt.imshow(image, cmap='gray')
    plt.title('Black with white square image')
    plt.axis('off')
    plt.show()
    
    # Display the gradient
    plt.figure(figsize=(6, 6))
    plt.imshow(gradient_magnitude, cmap='gray')
    plt.title('Gradient magnitude')
    plt.axis('off')
    plt.show()

    #Compute the gradient angle using Sobel
    gradient_angle = np.arctan2(sobely, sobelx)

    # Display the gradient angle image.
    plt.figure(figsize=(6, 6))
    plt.imshow(gradient_angle, cmap='gray') 
    plt.title('Gradient Angle Image')
    plt.axis('off')
    plt.show()

    #Horizontal
    mask_horizontal = np.array([[0, 1, 0],
                                [1, 0, 1],
                                [0, 1, 0]])

    #Vertical
    mask_vertical = np.array([[0, 1, 0],
                            [1, 0, 1],
                            [0, 1, 0]])

    # 45
    mask_45_degree = np.array([[1, 0, 0],
                            [0, 0, 0],
                            [0, 0, 1]])

    #-45
    mask_minus_45_degree = np.array([[0, 0, 1],
                                    [0, 0, 0],
                                    [1, 0, 0]])

    #Display
    plt.figure(figsize=(10, 5))
    plt.subplot(141), plt.imshow(mask_horizontal, cmap='gray'), plt.title('Horizontal Break')
    plt.axis('off')
    plt.subplot(142), plt.imshow(mask_vertical, cmap='gray'), plt.title('Vertical Break')
    plt.axis('off')
    plt.subplot(143), plt.imshow(mask_45_degree, cmap='gray'), plt.title('45° Break')
    plt.axis('off')
    plt.subplot(144), plt.imshow(mask_minus_45_degree, cmap='gray'), plt.title('-45° Break')
    plt.axis('off')
    plt.show()



# prob2()
prob3()