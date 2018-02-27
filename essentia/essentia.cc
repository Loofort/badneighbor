

#include <iostream>
#include <fstream>
#include <essentia/algorithmfactory.h>
//#include <essentia/essentiamath.h>
#include <essentia.h>

using namespace std;
using namespace essentia;
using namespace essentia::standard;

/*
//--------------------------------------------------------------
void SimpleEssentiaUtil::exit(){
    delete dcremoval;
    delete rms;
    delete energy;
    delete power;
    essentia::shutdown();
}

void SimpleEssentiaUtil::analyze(float * iBuffer, int bufferSize){
    
    vector <float> fBuffer;
    fBuffer.resize(bufferSize);
    memcpy(&fBuffer[0], iBuffer, sizeof(float) * bufferSize);
    for (int i=0; i<bufferSize;i++){
        audioBuffer[i] = (Real) fBuffer[i];
    }
    
    dcremoval->compute();
    rms->compute();
    energy->compute();
    power->compute();
    
    for (int i=0; i<spec.size(); i++)
        spectrum_f[i] = log10((float) spec[i]);

    rms_f = (float) rmsValue;
    
    energy_f = (float) energyValue;

    power_f = (float) powerValue;
}

void glutIdle()
{
    if (animate)
    {
        Pa_ReadStream( stream, sampleBlock, FRAMES_PER_BUFFER );
        for (int i = 0; i < FRAMES_PER_BUFFER; i++){
            buffer_to_analyze[i]	= sampleBlock[i];
        }
        audioAnalyzer.analyze(buffer_to_analyze, BUFFER_SIZE);
        // printf("read analyzed result: %f \n", audioAnalyzer.getRms());
        
        radius_rms =    0.2*global_scaling*audioAnalyzer.getRms();
        radius_energy = 0.2*global_scaling*audioAnalyzer.getEnergy();
        radius_power =  3.0*global_scaling*audioAnalyzer.getPower();
    }
    glutPostRedisplay();
}


void init_function()
{
    audioAnalyzer.setup(BUFFER_SIZE, SAMPLE_RATE);
    
    for(int i = 0; i < BUFFER_SIZE; i++) {
        sampleBlock[i]= 0.0;
        buffer_to_analyze[i]= 0.0;
    }
}
*/
//////////////////////////////////////////////////////////////////
AlgorithmFactory& initFactory() {
    essentia::init();
    return standard::AlgorithmFactory::instance();
}
AlgorithmFactory& factory = initFactory();
int sampleRate = 44100;
int frameSize = 2048;
int hopSize = 1024;

struct Analyzer {
    vector<Real> audioBuffer;
    vector<Real> audioBuffer_dc;

    Algorithm* dcremoval;
    Algorithm* energy;

    Real energyValue;
};

void* NewAnalyzer() {
    Analyzer* anl = new Analyzer();

    // register the algorithms in the factory(ies)
    anl->audioBuffer.resize(frameSize);

    anl->dcremoval = factory.create("DCRemoval", "sampleRate", sampleRate);
    anl->dcremoval->input("signal").set(anl->audioBuffer);
    anl->dcremoval->output("signal").set(anl->audioBuffer_dc);

    anl->energy = factory.create("Energy");
    anl->energy->input("array").set(anl->audioBuffer_dc);
    anl->energy->output("energy").set(anl->energyValue);
    return anl;
}

void DestroyAnalyzer(void *ptr) {
    Analyzer* anl = (Analyzer*)ptr;
    delete anl;
    essentia::shutdown();
}

ResultArr AnalyzeFile(void *ptr, const char *path) {
    Analyzer* anl = (Analyzer*)ptr;

    Algorithm* audio = factory.create("MonoLoader", "filename", path, "sampleRate", sampleRate);
    vector<Real> audioBuffer;
    audio->output("audio").set(audioBuffer);

    Algorithm* fc = factory.create("FrameCutter", "frameSize", frameSize, "hopSize", hopSize);
    fc->input("signal").set(audioBuffer);
    fc->output("frame").set(anl->audioBuffer);

    /*
    Algorithm* w = factory.create("Windowing", "type", "blackmanharris62");
    */

    audio->compute();

    Real* arr = 0;
    arr = (Real*)malloc((audioBuffer.size()/hopSize+2) * sizeof(Real)); 
    uint32_t cnt=0;

    while (true) {
        // compute a frame
        fc->compute();

        // if it was the last one (ie: it was empty), then we're done.
        if (!anl->audioBuffer.size()) {
        break;
        }

        // if the frame is silent, just drop it and go on processing
        //if (isSilent(anl->audioBuffer)) continue;

        anl->dcremoval->compute();
        anl->energy->compute();

        //pool.add("lowlevel.mfcc", anl->energyValue);
        //cout << anl->energyValue << " ";

        arr[cnt] = anl->energyValue;
        cnt++;
    }
    //cout << "\n";
    //cout << "ORIG: " << (audioBuffer.size()/hopSize+2) << "\n";
    //cout << "CNT: " << cnt << "\n";

    ResultArr result={0};
    result.Cnt = cnt;
    result.Res = arr;

    return result;
}