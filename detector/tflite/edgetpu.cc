#include <tensorflow/lite/experimental/c/c_api_internal.h>
#include <tensorflow/lite/interpreter.h>
#include <tensorflow/lite/kernels/register.h>
#include <tensorflow/lite/model.h>
#include <libedgetpu/edgetpu.h>

// Global TPU Instance
auto tpu_context = edgetpu::EdgeTpuManager::GetSingleton() -> NewEdgeTpuContext();

extern "C"
{
    // HasEdgeTPU returns if EdgeTPU is found
    int HasEdgeTPU() { return tpu_context != nullptr; }

    // RegisterEdgeTPUCustomOp returns the edgetpu custom op
    TfLiteRegistration *RegisterEdgeTPUCustomOp()
    {
        return edgetpu::RegisterCustomOp();
    }

    // EdgeTPUSetup sets the external context
    void EdgeTPUSetup(TFL_Interpreter *i, TFL_Model *m)
    {
        tflite::ops::builtin::BuiltinOpResolver resolver;
        resolver.AddCustom(edgetpu::kCustomOp, edgetpu::RegisterCustomOp());
        tflite::InterpreterBuilder(*m->impl, resolver)(&i->impl);
        i->impl->SetExternalContext(kTfLiteEdgeTpuContext, tpu_context.get());
    }
}
