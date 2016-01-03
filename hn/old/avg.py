#!/usr/bin/python2
# -*- coding: utf-8 -*-

import codecs
import sys
import re
from PyQt4 import QtGui


class AverageModificator(QtGui.QMainWindow):

    def __init__(self):
        super(AverageModificator, self).__init__()
        self.initUI()

    def initUI(self):
        vbox = QtGui.QVBoxLayout()

        self.mainWidget = QtGui.QWidget(self)
        self.mainWidget.setLayout(vbox)

        self.inputFileSelector = QtGui.QPushButton("Open file")
        self.inputFileSelector.clicked.connect(self.inputSelectFile)
        self.outputFileSelector = QtGui.QPushButton("Save As")
        self.outputFileSelector.clicked.connect(self.outputSaveFile)
        self.execTask = QtGui.QPushButton("Execute")
        self.execTask.clicked.connect(self.doTask)

        vbox.addWidget(self.inputFileSelector)
        vbox.addWidget(self.outputFileSelector)
        vbox.addWidget(self.execTask)

        self.setCentralWidget(self.mainWidget)

        self.setGeometry(300, 300, 250, 200)
        self.setWindowTitle('Average Modificator')
        self.show()

        self.inputFile = '/home/post-l/Downloads/OTTO.TXT'
        self.outputFile = '/home/post-l/Downloads/test.txt'

    def inputSelectFile(self):
        self.inputFile = QtGui.QFileDialog.getOpenFileName(self, 'Open file', selectedFilter='*.txt')

    def outputSaveFile(self):
        self.outputFile = QtGui.QFileDialog.getSaveFileName(self, 'Save As', selectedFilter='*.txt')

    def doTask(self):
        if not self.inputFile:
            QtGui.QMessageBox.critical(self, 'Error', 'Select an input file')
            return
        if not self.outputFile:
            QtGui.QMessageBox.critical(self, 'Error', 'Select an output file')
            return
        interval, ok = QtGui.QInputDialog.getInt(self, 'Set Interval', 'Interval', 12, 1)
        if not ok:
            return
        output = ""
        with open(self.inputFile, 'r') as f:
            i = 0
            pattern = re.compile(r'(\d+)-(\d+)-(\d+) (\d+):(\d+):(\d+) *')
            for line in f:
                if pattern.match(line):
                    t = line.split()
                    if i == 0:
                        val = float(t[2])
                        firstDate = (t[0], t[1])
                    else:
                        val = val + float(t[2]) / 2.0
                    i = i + 1
                    if i == interval:
                        output += "%s %s %.1f\r\n" % (firstDate[0], firstDate[1], val)
                        i = 0
                else:
                    if i != 0:
                        output += "%s %s %.1f\r\n" % (firstDate[0], firstDate[1], val)
                        i = 0
                    output += line.decode('latin1')
            f.close()
        with codecs.open(self.outputFile, 'w', 'utf-8') as f:
            f.write(output)
            f.close()
        QtGui.QMessageBox.information(self, 'Success', 'You\'re file has been created!')

def main():
    app = QtGui.QApplication(sys.argv)
    am = AverageModificator()
    sys.exit(app.exec_())


if __name__ == '__main__':
    main()
