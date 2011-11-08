#include <math.h>
#include <GL/glut.h>
#include <QtGui>
#include <QtOpenGL>
#include "glwidget.h"
#include <iostream>
#include "container.h"

#define PI (3.141592653589793)

#ifndef GL_MULTISAMPLE
#define GL_MULTISAMPLE  0x809D
#endif

GLWidget::GLWidget(QWidget *parent)
  : QGLWidget(QGLFormat(QGL::SampleBuffers), parent)
{
  xRot = 0;
  yRot = 0;
  zRot = 0;
  usePtr = false;
  displayOn = false;

  qtGreen  = QColor::fromCmykF(0.40, 0.0, 1.0, 0.0);
  qtPurple = QColor::fromCmykF(0.39, 0.39, 0.0, 0.0);
}

GLWidget::~GLWidget()
{
}

void GLWidget::updateData(array_ptr data)
{
  dataPtr    = data;
  displayOn  = true;
  
  // Update the display
  updateCOM();
  updateExtent();
  updateGL();
}

QSize GLWidget::minimumSizeHint() const
{
  return QSize(200, 200);
}

QSize GLWidget::sizeHint() const
{
  return QSize(800, 400);
}

static void qNormalizeAngle(int &angle)
{
  while (angle < 0)
    angle += 360 * 16;
  while (angle > 360 * 16)
    angle -= 360 * 16;
}

void GLWidget::setXRotation(int angle)
{
  qNormalizeAngle(angle);
  if (angle != xRot) {
    xRot = angle;
    emit xRotationChanged(angle);
    updateGL();
  }
}

void GLWidget::setYRotation(int angle)
{
  qNormalizeAngle(angle);
  if (angle != yRot) {
    yRot = angle;
    emit yRotationChanged(angle);
    updateGL();
  }
}

void GLWidget::setZRotation(int angle)
{
  qNormalizeAngle(angle);
  if (angle != zRot) {
    zRot = angle;
    emit zRotationChanged(angle);
    updateGL();
  }
}

void GLWidget::setXSliceLow(int low)
{
  if (xSliceLow != low) {
    xSliceLow = low;
    updateGL();
  }
}

void GLWidget::setXSliceHigh(int high)
{
  if (xSliceLow != high) {
    xSliceHigh = high;
    updateGL();
  }
}

void GLWidget::setYSliceLow(int low)
{
  if (ySliceLow != low) {
    ySliceLow = low;
    updateGL();
  }
}

void GLWidget::setYSliceHigh(int high)
{
  if (ySliceLow != high) {
    ySliceHigh = high;
    updateGL();
  }
}

void GLWidget::setZSliceLow(int low)
{
  if (zSliceLow != low) {
    zSliceLow = low;
    updateGL();
  }
}

void GLWidget::setZSliceHigh(int high)
{
  if (zSliceLow != high) {
    zSliceHigh = high;
    updateGL();
  }
}

void GLWidget::updateCOM()
{
  const long unsigned int *size = dataPtr->shape();
  xcom = (float)size[0]*0.5;
  ycom = (float)size[1]*0.5;
  zcom = (float)size[2]*0.5;
}

void GLWidget::updateExtent() 
{
  const long unsigned int *size = dataPtr->shape();
  xmax = size[0];
  ymax = size[1];
  zmax = size[2];
  xmin = 0.0;
  ymin = 0.0;
  zmin = 0.0;
}

void GLWidget::initializeGL()
{
  // GLUT wants argc and argv... qt obscures these in the class
  // so let us circumenvent this problem...
  int argc = 1;
  const char* argv[] = {"Sloppy","glut"};
  glutInit(&argc, (char**)argv);

  qglClearColor(qtPurple.dark());

  glHint(GL_PERSPECTIVE_CORRECTION_HINT, GL_NICEST);
  glColorMaterial ( GL_FRONT, GL_AMBIENT_AND_DIFFUSE);
  glEnable(GL_COLOR_MATERIAL);
  glEnable(GL_DEPTH_TEST);
  glEnable(GL_CULL_FACE);
  glShadeModel(GL_SMOOTH);
  glEnable(GL_LIGHTING);
  glEnable(GL_LIGHT0);
  glEnable(GL_MULTISAMPLE);
  static GLfloat lightPosition[4] = { 0.5, 10.0, 7.0, 1.0 };
  glLightfv(GL_LIGHT0, GL_POSITION, lightPosition);

  // Display List for cone
  cone = glGenLists(1);
  // Draw a cone pointing along the z axis
  glNewList(cone, GL_COMPILE);
    glPushMatrix();
    //glRotatef(0.0f,0.0f,0.0f,0.0f);
    glutSolidCone(0.2f, 0.7f, 10, 1);
    glPopMatrix();
  glEndList();

  // Set the slice initial conditions
  xSliceLow=ySliceLow=zSliceLow=0;
  xSliceHigh=ySliceHigh=zSliceHigh=16*100;

  // Initial view
  zoom=0.5;
}

void GLWidget::paintGL()
{
  glClear(GL_COLOR_BUFFER_BIT | GL_DEPTH_BUFFER_BIT);
  glLoadIdentity();
  glTranslatef(0.0, 0.0, -10.0 + zoom);
  glRotatef(xRot / 16.0, 1.0, 0.0, 0.0);
  glRotatef(yRot / 16.0, 0.0, 1.0, 0.0);
  glRotatef(zRot / 16.0, 0.0, 0.0, 1.0);
  //std::cout << "Zoom: " << zoom << std::endl;

  if (displayOn) {
    const long unsigned int *size = dataPtr->shape();
    int xnodes = size[0];
    int ynodes = size[1];
    int znodes = size[2];
    float theta, phi, mag;

    for(int i=0; i<xnodes; i++)
      {
	for(int j=0; j<ynodes; j++)
	  {
	    for(int k=0; k<znodes; k++)
	      {
		mag = sqrt( (*dataPtr)[i][j][k][0] * (*dataPtr)[i][j][k][0] +
			    (*dataPtr)[i][j][k][1] * (*dataPtr)[i][j][k][1] +
			    (*dataPtr)[i][j][k][2] * (*dataPtr)[i][j][k][2]);
		if (mag > 1.0 &&
		    i >= (xmax-xmin)*(float)xSliceLow/1600.0 &&
		    i <= (xmax-xmin)*(float)xSliceHigh/1600.0 &&
		    j >= (ymax-ymin)*(float)ySliceLow/1600.0 &&
		    j <= (ymax-ymin)*(float)ySliceHigh/1600.0 &&
		    k >= (zmax-zmin)*(float)zSliceLow/1600.0 &&
		    k <= (zmax-zmin)*(float)zSliceHigh/1600.0) 
		  {
		    
		    theta = acos(  (*dataPtr)[i][j][k][2]/mag);
		    phi   = atan2( (*dataPtr)[i][j][k][1],  (*dataPtr)[i][j][k][0]);
		    
		    glPushMatrix();
		    glTranslatef((float)i-xcom,(float)j-ycom, (float)k-zcom);
		    glColor3f(sin(phi), cos(phi), cos(phi+1.0f));
		    
		    glRotatef(180.0*theta/PI, 0.0, 1.0, 0.0);
		    glRotatef(180.0*phi/PI,   0.0, 0.0, 1.0);
		    
		    glCallList(cone);
		    glPopMatrix();
		  }
	      }
	  }
      }
  } 
}

void GLWidget::resizeGL(int width, int height)
{
  glMatrixMode(GL_PROJECTION);
  glLoadIdentity();
  gluPerspective(60.0, (float)width / (float)height, 0.1, 80.0);
  glMatrixMode(GL_MODELVIEW);
  glViewport(0, 0, width, height);
}

void GLWidget::mousePressEvent(QMouseEvent *event)
{
  lastPos = event->pos();
}

void GLWidget::mouseMoveEvent(QMouseEvent *event)
{
  int dx = event->x() - lastPos.x();
  int dy = event->y() - lastPos.y();

  if (event->buttons() & Qt::LeftButton) {
    setXRotation(xRot + 8 * dy);
    setYRotation(yRot + 8 * dx);
  } else if (event->buttons() & Qt::RightButton) {
    setXRotation(xRot + 8 * dy);
    setZRotation(zRot + 8 * dx);
  }
  lastPos = event->pos();
}

void GLWidget::wheelEvent(QWheelEvent *event)
{
    if(event->orientation() == Qt::Vertical)
      {
	zoom += (float)(event->delta()) / 100;
	updateGL();
      }
}
