import cv2
import numpy as np
import matplotlib.pyplot as plt
from scipy.fft import fft2, ifft2, fftshift, ifftshift
from scipy.signal import wiener

def blurringFilter(shape, k=0.0005):
    #Create MeshGrid
    u = np.arange(-shape[0]//2, shape[0]//2)
    v = np.arange(-shape[1]//2, shape[1]//2)
    U, V = np.meshgrid(u, v)

    #Center is 0,0
    u0 = v0 = 0

    #Constructing the filter
    D_squared = (U - u0)**2 + (V - v0)**2
    filter = np.exp(-k * (D_squared ** (5/6)))
    #Shift to center
    return np.fft.ifftshift(filter)  

#FFT
def freqFilter(img, givenFilter):
    #Frequency domain
    F = fft2(img)
    #Shift to center
    F_shifted = fftshift(F)  
    # Apply filter
    G = givenFilter * F_shifted
    G_shifted = fftshift(G)  # Shift back
    # Convert back to spatial domain
    spatial = ifft2(G_shifted)
    spatial = np.abs(spatial)
    spatial = cv2.normalize(spatial, None, alpha=0, beta=255, norm_type=cv2.NORM_MINMAX)  # Normalize to [0, 255]
    return np.uint8(spatial)

def gaussian(img, mean=0, sigma=0.1):
    noise = np.random.normal(mean, sigma, img.shape)
    noisyImg = img + noise
    return np.uint8(np.clip(noisyImg, 0, 255)) #Renormalize


def makeHisto(img, title):
    plt.hist(img.ravel(), 256, [0, 1])
    plt.title(title)
    plt.show()

def exponentialNoise(img, scale=1/0.1):
    exponential = np.random.exponential(scale, img.shape)
    noisyImage = img + exponential
    #We have to make the values between 0 and 1 again
    noisyImage = np.clip(noisyImage, 0, 1)  
    return noisyImage

def uniformNoise(img, low=0, high=0.5):
    noise = np.random.uniform(low, high, img.shape)
    noiseImage = img + noise
    #We have to make the values between 0 and 1 again
    noiseImage = np.clip(noiseImage, 0, 1)  
    return noiseImage

def gaussianNoiseProb1(img, mean=0, std=0.1):
    gaussian = np.random.normal(mean, std, img.shape)
    noisyImage = img + gaussian
    #We have to make the values between 0 and 1 again
    noisyImage = np.clip(noisyImage, 0, 1) 
    return noisyImage

def prob1():
    img = cv2.imread("DIP3E_CH05_Original_Images/FigP0528(c)(doughnut).tif", 0) 
    #Make the image values from 0 to 1
    img = img / 255

    #apply our noise filters
    gaussianImg = gaussianNoiseProb1(img)
    exponentialImg = exponentialNoise(img)
    uniformImg = uniformNoise(img)

    #Create Histograms
    makeHisto(img, 'Original Image Histogram')
    makeHisto(gaussianImg, 'Gaussian Noise Histogram')
    makeHisto(exponentialImg, 'Exponential Noise Histogram')
    makeHisto(uniformImg, 'Uniform Noise Histogram')

    #Display
    cv2.imshow('Original Image', img)
    cv2.imshow('Gaussian Noisy Image', gaussianImg)
    cv2.imshow('Exponential Noisy Image', exponentialImg)
    cv2.imshow('Uniform Noisy Image', uniformImg)
    cv2.waitKey(0)
    cv2.destroyAllWindows()

    
    
def prob2():
    img = cv2.imread("DIP3E_CH05_Original_Images/Fig0525(a)(aerial_view_no_turb).tif", 0) 
    cv2.imshow("test",img) #testing
    #Seperate Blurring filter function
    blurredFilter = blurringFilter(img.shape)
    cv2.imshow("Blurring Filter:", blurredFilter)

    #Apply filter
    blurredImg = freqFilter(img, blurredFilter)
    cv2.imshow("Blurred Image", blurredImg)

    #Add Gaussian 
    noisyImg = gaussian(img)
    cv2.imshow("Gaussian Image", noisyImg)
    
    #As far as i know you need the variance for weiner filter in python
    noise_variance = np.var(noisyImg - img)

    #Wiener filter 
    for K in [0.01, 0.1, 1]:
        wienerImg = wiener(noisyImg, noise=noise_variance / K)
        #Renormalize with these next two
        wienerImg = np.clip(wienerImg, 0, 255) 
        wienerImg = np.uint8(wienerImg) 
        cv2.imshow(f"Wiener Filtered Image K={K}", wienerImg)

    cv2.waitKey(0)
    cv2.destroyAllWindows()

def prob3():
    img = cv2.imread("DIP3E_CH05_Original_Images/FigP0501(filtering).tif", 0) 
    #Using a cross square.
    SE = np.array([[0, 1, 0],
                   [1, 1, 1],
                   [0, 1, 0]], dtype=np.uint8)

    #Erosion
    erodedImg = cv2.erode(img, SE)

    #Dilation
    dilatedImg = cv2.dilate(img, SE)

    #Display
    cv2.imshow("Original", img)
    cv2.imshow("Eroded", erodedImg)
    cv2.imshow("Dilated", dilatedImg)
    cv2.waitKey(0)
    cv2.destroyAllWindows()

#Uncomment to run 
prob1()
# prob2()
# prob3()