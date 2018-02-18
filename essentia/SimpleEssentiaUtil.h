
#ifndef __SimpleAudioBasedMIRVisualizationTemplate__SimpleEssentiaUtil__
#define __SimpleAudioBasedMIRVisualizationTemplate__SimpleEssentiaUtil__


#include <iostream>
#include "essentia/algorithmfactory.h"
#include "essentia/essentiamath.h"
#include "essentia/pool.h"

using namespace essentia;
using namespace standard;
using namespace std;

class SimpleEssentiaUtil
{
public:
    
    void setup(int bufferSize, int sampleRate);
    void exit();
    void analyze(float * iBuffer, int bufferSize);
    
    float getRms(){return rms_f;}
    float getEnergy(){return energy_f;}
    float getPower(){return power_f;}
    
private:
    
    int sr;
    vector<float> spectrum_f;
    float rms_f, energy_f, power_f;
    
    Algorithm* spectrum;
    Algorithm* rms;
    Algorithm* energy;
    Algorithm* power;
    Algorithm* dcremoval;
    
    Real rmsValue;
    Real powerValue;
    Real energyValue;
    
    vector<Real> spec;
    vector<Real> audioBuffer;
    vector<Real> audioBuffer_dc;
};



#endif /* defined(__SimpleAudioBasedMIRVisualizationTemplate__SimpleEssentiaUtil__) */