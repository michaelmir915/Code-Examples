##Some small examples of functions from my research:

from matplotlib import pyplot as plt
from scipy.io import loadmat
from scipy.io import *
from io import StringIO
from scipy import signal
from scipy import pi
from scipy.signal import argrelmax
from scipy.fft import *
# import pandas as pd
from numpy import abs, linspace
import numpy as np
from PIL import Image
import multiprocessing as mp
from scipy.ndimage import median_filter, gaussian_filter

def extractData():
    data = loadmat('RFData.mat')
    preCompressed = data['PreCompressedRFData']
    postCompressed = data['PostCompressedRFData']
    overlapDATA = [[]]
    noOverlapDATA = [[]]

    for colNumber in range(128):
    # for colNumber in range(8):
        print("COLUMN NUMBER: " + str(colNumber))
        for k in data.keys():
            if k.startswith('PreCompressedRFData'):
                print(k + " " + data[k].dtype.name + " " + str(data[k].shape))
            elif k.startswith('PostCompressedRFData'):
                print(k + " " + data[k].dtype.name + " " + str(data[k].shape))
        preColumn = []
        postColumn = []
        # print(preCompressed[0][129])
        # x_axis = list(range(0,40,.1039))
        x_axis =[]
        x_axis_CrossCorr = []
        for i in range(4156):
            preColumn.append(preCompressed[i][colNumber])
            postColumn.append(postCompressed[i][colNumber])
            x_axis.append((40/4156)*i)
            x_axis_CrossCorr.append((40/4156)*i)
        plt.plot(x_axis, preColumn, label = "Pre-Compressed")
        plt.plot(x_axis, postColumn, label = "Post-Compressed")
        plt.title("Pre Column and Post Column")
        plt.legend()
        # plt.show()


        crossCorr = signal.correlate(preColumn, postColumn, mode='full')
        crossCorrX = linspace(0,80, 8311)
        plt.plot(crossCorrX, crossCorr)
        plt.title("Cross correlation pre and post column")
        # plt.show() //uncomment me
        plt.close('all')

    ##Windows of ~2mm each PRE-COMPRESSION
    def noOverlap(noOverlapPre):
        noOverlapPre = []
        fullNoOvPreList = []
        counter = 0
        for i in range(4156):
            if (((i % 200 == 0) and i !=0)):
                TwentyList = linspace(counter,counter+2,200)
                counter += 2
                plt.plot(TwentyList, noOverlapPre)
                # plt.show() ##uncomment to see first 19 windows
                fullNoOvPreList.append(noOverlapPre)
                noOverlapPre = []
            noOverlapPre.append(preColumn[i])
        fullNoOvPreList.append(noOverlapPre)



    ##Windows of ~2mm each with 40% overlap (0.8mm) PRE-COMPRESSION
    def overlappedData(overlapPreList):
        overlapPre = []
        fullOverlapPreList = []
        pixelNo = 0
        counter = pixelNo + 200
        temp = 0
        overlapPre = []
        while (pixelNo < 4156):
            if (counter == pixelNo):
                overlapPreList = linspace(temp, temp+2, 200)
                pixelNo -= 80
                counter = pixelNo + 200
                plt.plot(overlapPreList, overlapPre)
                plt.plot()
                temp += 2
                fullOverlapPreList.append(overlapPre)
                overlapPre = []
            overlapPre.append(preColumn[pixelNo])
            pixelNo += 1
        fullOverlapPreList.append(overlapPre)
        plt.title("40% Overlap Pre-Compression Windows")
        plt.show()


    ##Windows of ~2mm each
    plt.close('all')


    #No Overlap Cross Correlation
    def crossCorr(preList,postList):
        maxX = -1
        maxY = -1
        peakListY = []
        peakListX = []
        fullNoCrossCorrList = []
        for i in range(len(preList)):
            noOverlapCrossCorr = signal.correlate(preList[i], postList[i], mode='full')
            newCrossCorrX = linspace(i*4, (i+1)*4, len(noOverlapCrossCorr))
            plt.plot(newCrossCorrX, noOverlapCrossCorr)
            fullNoCrossCorrList.append(noOverlapCrossCorr)
            if (max(noOverlapCrossCorr) > maxY):
                maxY = max(noOverlapCrossCorr)
                maxX = newCrossCorrX[np.where(noOverlapCrossCorr == maxY)]
                peakListY.append(maxY)
                peakListX.append(maxX)
                maxY = -1
                maxX = -1
        crossCorrX = linspace(0,80, 8311)
        plt.title("Cross correlation no overlap")
        # plt.show() 
        plt.close('all')
        plt.title("all peaks NO OVERLAP")
        # plt.plot(peakListX, peakListY)
        xAxisNum3 = linspace (0,21,21)
        moddedXAxis = []
        for i in range(len(peakListX)):
            moddedXAxis.append(peakListX[i] % 4)
        # print(moddedXAxis)
        noOverlapDATA.append(moddedXAxis)
        plt.plot(xAxisNum3, moddedXAxis, marker='D')
        plt.xlabel("Segment Number")
        plt.ylabel("Distance to peak (mm)")
        # plt.show()
        plt.close('all')
        ##graph should be x-axis(0,1,2,3,4,5,6,7...)
        ##y-axis (2.02, 6.07,10.1,14.15,18.19)

        ### With Overlap Cross Correlation
        maxX = -1
        maxY = -1
        peakListY = []
        peakListX = []
        for i in range(len(preList)):
            overlapCrossCorr = signal.correlate(preList[i], postList[i], mode='full')
            newCrossCorrX = linspace(i*4, (i+1)*4, len(overlapCrossCorr))
            plt.plot(newCrossCorrX, overlapCrossCorr)
            if (max(overlapCrossCorr) > maxY):
                maxY = max(overlapCrossCorr)
                maxX = newCrossCorrX[np.where(overlapCrossCorr == maxY)]
                peakListY.append(maxY)
                peakListX.append(maxX)
                maxY = -1
                maxX = -1
        plt.title("Cross correlation with overlap")
        # print("The coordinates of the peak is: ", maxX, ",", maxY,".")
        # plt.show()
        plt.close('all')
        plt.title("all peaks WITH OVERLAP")
        # plt.plot(peakListX, peakListY)
        xAxisNum3 = linspace (0,34,34)
        moddedXAxis = []
        for i in range(len(peakListX)):
            moddedXAxis.append(peakListX[i] % 4)
        # print(moddedXAxis)
        overlapDATA.append(moddedXAxis)
        # print(moddedXAxis)
        plt.plot(xAxisNum3, moddedXAxis, marker='D')
        plt.xlabel("Segment Number")
        plt.ylabel("Distance to peak (mm)")
        # plt.show()
        plt.close('all')